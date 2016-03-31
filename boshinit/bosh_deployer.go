package boshinit

import (
	"fmt"

	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit/manifests"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
)

type BOSHDeployer struct {
	manifestBuilder boshInitManifestBuilder
	runner          boshInitRunner
	logger          logger
}

type BOSHDeployOutput struct {
	Credentials        map[string]string
	BOSHInitState      State
	DirectorSSLKeyPair ssl.KeyPair
	BOSHInitManifest   string
}

type logger interface {
	Step(message string)
	Println(string)
}

type boshInitManifestBuilder interface {
	Build(manifests.ManifestProperties) (manifests.Manifest, manifests.ManifestProperties, error)
}

type boshInitRunner interface {
	Deploy(manifest []byte, privateKey string, state State) (State, error)
}

func NewBOSHDeployer(manifestBuilder boshInitManifestBuilder, runner boshInitRunner, logger logger) BOSHDeployer {
	return BOSHDeployer{
		manifestBuilder: manifestBuilder,
		runner:          runner,
		logger:          logger,
	}
}

func (b BOSHDeployer) Deploy(input BOSHDeployInput) (BOSHDeployOutput, error) {
	manifest, manifestProperties, err := b.manifestBuilder.Build(manifests.ManifestProperties{
		DirectorUsername: input.DirectorUsername,
		DirectorPassword: input.DirectorPassword,
		SubnetID:         input.InfrastructureConfiguration.SubnetID,
		AvailabilityZone: input.InfrastructureConfiguration.AvailabilityZone,
		ElasticIP:        input.InfrastructureConfiguration.ElasticIP,
		AccessKeyID:      input.InfrastructureConfiguration.AccessKeyID,
		SecretAccessKey:  input.InfrastructureConfiguration.SecretAccessKey,
		SecurityGroup:    input.InfrastructureConfiguration.SecurityGroup,
		Region:           input.InfrastructureConfiguration.AWSRegion,
		DefaultKeyName:   input.EC2KeyPair.Name,
		SSLKeyPair:       input.SSLKeyPair,
		Credentials:      manifests.NewInternalCredentials(input.Credentials),
	})
	if err != nil {
		return BOSHDeployOutput{}, err
	}

	yaml, err := candiedyaml.Marshal(manifest)
	if err != nil {
		return BOSHDeployOutput{}, err
	}

	state, err := b.runner.Deploy(yaml, input.EC2KeyPair.PrivateKey, input.State)
	if err != nil {
		return BOSHDeployOutput{}, err
	}

	b.logger.Println(fmt.Sprintf("Director Address:  %s", manifestProperties.ElasticIP))
	b.logger.Println("Director Username: " + manifestProperties.DirectorUsername)
	b.logger.Println("Director Password: " + manifestProperties.DirectorPassword)

	return BOSHDeployOutput{
		BOSHInitState:      state,
		DirectorSSLKeyPair: manifestProperties.SSLKeyPair,
		Credentials:        manifestProperties.Credentials.ToMap(),
		BOSHInitManifest:   string(yaml),
	}, nil
}
