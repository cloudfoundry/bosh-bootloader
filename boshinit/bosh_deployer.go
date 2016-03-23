package boshinit

import (
	"fmt"

	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
)

type BOSHDeployer struct {
	manifestBuilder boshInitManifestBuilder
	runner          boshInitRunner
	logger          logger
}

type BOSHDeployInput struct {
	DirectorUsername string
	DirectorPassword string
	State            State
	Stack            cloudformation.Stack
	AWSRegion        string
	SSLKeyPair       ssl.KeyPair
	EC2KeyPair       ec2.KeyPair
	Credentials      InternalCredentials
}

type BOSHDeployOutput struct {
	Credentials        InternalCredentials
	BOSHInitState      State
	DirectorSSLKeyPair ssl.KeyPair
}

type boshInitManifestBuilder interface {
	Build(ManifestProperties) (Manifest, ManifestProperties, error)
}

type boshInitRunner interface {
	Deploy(manifest, privateKey []byte, state State) (State, error)
}

func NewBOSHDeployer(manifestBuilder boshInitManifestBuilder, runner boshInitRunner, logger logger) BOSHDeployer {
	return BOSHDeployer{
		manifestBuilder: manifestBuilder,
		runner:          runner,
		logger:          logger,
	}
}

func (b BOSHDeployer) Deploy(input BOSHDeployInput) (BOSHDeployOutput, error) {
	manifest, manifestProperties, err := b.manifestBuilder.Build(ManifestProperties{
		DirectorUsername: input.DirectorUsername,
		DirectorPassword: input.DirectorPassword,
		SubnetID:         input.Stack.Outputs["BOSHSubnet"],
		AvailabilityZone: input.Stack.Outputs["BOSHSubnetAZ"],
		ElasticIP:        input.Stack.Outputs["BOSHEIP"],
		AccessKeyID:      input.Stack.Outputs["BOSHUserAccessKey"],
		SecretAccessKey:  input.Stack.Outputs["BOSHUserSecretAccessKey"],
		SecurityGroup:    input.Stack.Outputs["BOSHSecurityGroup"],
		Region:           input.AWSRegion,
		DefaultKeyName:   input.EC2KeyPair.Name,
		SSLKeyPair:       input.SSLKeyPair,
		Credentials:      input.Credentials,
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

	b.logger.Println(fmt.Sprintf("Director Address:  https://%s:25555", manifestProperties.ElasticIP))
	b.logger.Println("Director Username: " + manifestProperties.DirectorUsername)
	b.logger.Println("Director Password: " + manifestProperties.DirectorPassword)

	return BOSHDeployOutput{
		BOSHInitState:      state,
		DirectorSSLKeyPair: manifestProperties.SSLKeyPair,
		Credentials:        manifestProperties.Credentials,
	}, nil
}
