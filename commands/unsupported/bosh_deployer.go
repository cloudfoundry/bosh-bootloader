package unsupported

import (
	"fmt"

	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
)

type BOSHDeployer struct {
	manifestBuilder boshInitManifestBuilder
	runner          boshInitRunner
	logger          logger
}

type BOSHDeployInput struct {
	State      boshinit.State
	Stack      cloudformation.Stack
	AWSRegion  string
	SSLKeyPair ssl.KeyPair
	EC2KeyPair ec2.KeyPair
}

type boshInitManifestBuilder interface {
	Build(boshinit.ManifestProperties) (boshinit.Manifest, boshinit.ManifestProperties, error)
}

type boshInitRunner interface {
	Deploy(manifest, privateKey []byte, state boshinit.State) (boshinit.State, error)
}

type logger interface {
	Println(string)
}

func NewBOSHDeployer(manifestBuilder boshInitManifestBuilder, runner boshInitRunner, logger logger) BOSHDeployer {
	return BOSHDeployer{
		manifestBuilder: manifestBuilder,
		runner:          runner,
		logger:          logger,
	}
}

func (b BOSHDeployer) Deploy(input BOSHDeployInput) (boshinit.State, ssl.KeyPair, error) {
	manifest, manifestProperties, err := b.manifestBuilder.Build(boshinit.ManifestProperties{
		DirectorUsername: "admin",
		DirectorPassword: "admin",
		SubnetID:         input.Stack.Outputs["BOSHSubnet"],
		AvailabilityZone: input.Stack.Outputs["BOSHSubnetAZ"],
		ElasticIP:        input.Stack.Outputs["BOSHEIP"],
		AccessKeyID:      input.Stack.Outputs["BOSHUserAccessKey"],
		SecretAccessKey:  input.Stack.Outputs["BOSHUserSecretAccessKey"],
		SecurityGroup:    input.Stack.Outputs["BOSHSecurityGroup"],
		Region:           input.AWSRegion,
		DefaultKeyName:   input.EC2KeyPair.Name,
		SSLKeyPair:       input.SSLKeyPair,
	})
	if err != nil {
		return boshinit.State{}, ssl.KeyPair{}, err
	}

	yaml, err := candiedyaml.Marshal(manifest)
	if err != nil {
		return boshinit.State{}, ssl.KeyPair{}, err
	}

	state, err := b.runner.Deploy(yaml, input.EC2KeyPair.PrivateKey, input.State)
	if err != nil {
		return boshinit.State{}, ssl.KeyPair{}, err
	}

	b.logger.Println(fmt.Sprintf("Director Address:  https://%s:25555", manifestProperties.ElasticIP))
	b.logger.Println("Director Username: " + manifestProperties.DirectorUsername)
	b.logger.Println("Director Password: " + manifestProperties.DirectorPassword)

	return state, manifestProperties.SSLKeyPair, nil
}
