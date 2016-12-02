package boshinit

import (
	"github.com/cloudfoundry/bosh-bootloader/boshinit/manifests"
	"gopkg.in/yaml.v2"
)

const (
	BOSH_BOOTLOADER_COMMON_NAME = "BOSH Bootloader"
)

type Executor struct {
	manifestBuilder manifestBuilder

	deployCommand command
	deleteCommand command
	logger        logger
}

type logger interface {
	Step(message string, a ...interface{})
	Println(string)
}

type manifestBuilder interface {
	Build(string, manifests.ManifestProperties) (manifests.Manifest, manifests.ManifestProperties, error)
}

type command interface {
	Execute(manifest []byte, privateKey string, state State) (State, error)
}

func NewExecutor(manifestBuilder manifestBuilder, deployCommand command, deleteCommand command, logger logger) Executor {
	return Executor{
		manifestBuilder: manifestBuilder,
		deployCommand:   deployCommand,
		deleteCommand:   deleteCommand,
		logger:          logger,
	}
}

func (e Executor) Delete(boshInitManifest string, boshInitState State, ec2PrivateKey string) error {
	e.logger.Step("destroying bosh director")

	_, err := e.deleteCommand.Execute([]byte(boshInitManifest), ec2PrivateKey, boshInitState)
	if err != nil {
		return err
	}

	return nil
}

func (e Executor) Deploy(input DeployInput) (DeployOutput, error) {
	manifest, manifestProperties, err := e.manifestBuilder.Build(input.IAAS, manifests.ManifestProperties{
		DirectorName:     input.DirectorName,
		DirectorUsername: input.DirectorUsername,
		DirectorPassword: input.DirectorPassword,
		SubnetID:         input.InfrastructureConfiguration.SubnetID,
		AvailabilityZone: input.InfrastructureConfiguration.AvailabilityZone,
		CACommonName:     BOSH_BOOTLOADER_COMMON_NAME,
		ElasticIP:        input.InfrastructureConfiguration.ElasticIP,
		AccessKeyID:      input.InfrastructureConfiguration.AccessKeyID,
		SecretAccessKey:  input.InfrastructureConfiguration.SecretAccessKey,
		SecurityGroup:    input.InfrastructureConfiguration.SecurityGroup,
		Region:           input.InfrastructureConfiguration.AWSRegion,
		DefaultKeyName:   input.EC2KeyPair.Name,
		SSLKeyPair:       input.SSLKeyPair,
		Credentials:      manifests.NewInternalCredentials(input.Credentials),
		GCP: manifests.ManifestPropertiesGCP{
			Zone:           input.InfrastructureConfiguration.GCP.Zone,
			NetworkName:    input.InfrastructureConfiguration.GCP.NetworkName,
			SubnetworkName: input.InfrastructureConfiguration.GCP.SubnetworkName,
			BOSHTag:        input.InfrastructureConfiguration.GCP.BOSHTag,
			InternalTag:    input.InfrastructureConfiguration.GCP.InternalTag,
			Project:        input.InfrastructureConfiguration.GCP.Project,
			JsonKey:        input.InfrastructureConfiguration.GCP.JsonKey,
		},
	})
	if err != nil {
		return DeployOutput{}, err
	}

	manifestYAML, err := yaml.Marshal(manifest)
	if err != nil {
		return DeployOutput{}, err
	}

	e.logger.Step("deploying bosh director")
	state, err := e.deployCommand.Execute(manifestYAML, input.EC2KeyPair.PrivateKey, input.State)
	if err != nil {
		return DeployOutput{}, err
	}

	return DeployOutput{
		BOSHInitState:      state,
		DirectorSSLKeyPair: manifestProperties.SSLKeyPair,
		Credentials:        manifestProperties.Credentials.ToMap(),
		BOSHInitManifest:   string(manifestYAML),
	}, nil
}
