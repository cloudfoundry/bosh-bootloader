package unsupported

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
)

type BOSHDeployer struct {
	stackManager    stackManager
	manifestBuilder boshInitManifestBuilder
}

type boshInitManifestBuilder interface {
	Build(boshinit.ManifestProperties) (boshinit.Manifest, error)
}

func NewBOSHDeployer(stackManager stackManager, manifestBuilder boshInitManifestBuilder) BOSHDeployer {
	return BOSHDeployer{
		stackManager:    stackManager,
		manifestBuilder: manifestBuilder,
	}
}

func (b BOSHDeployer) Deploy(cloudformationClient cloudformation.Client, region string, keyPairName string, sslKeyPair ssl.KeyPair) (ssl.KeyPair, error) {
	stack, err := b.stackManager.Describe(cloudformationClient, STACKNAME)
	if err != nil {
		return ssl.KeyPair{}, err
	}

	manifestProperties := boshinit.ManifestProperties{
		SubnetID:         stack.Outputs["BOSHSubnet"],
		AvailabilityZone: stack.Outputs["BOSHSubnetAZ"],
		ElasticIP:        stack.Outputs["BOSHEIP"],
		AccessKeyID:      stack.Outputs["BOSHUserAccessKey"],
		SecretAccessKey:  stack.Outputs["BOSHUserSecretAccessKey"],
		SecurityGroup:    stack.Outputs["BOSHSecurityGroup"],
		Region:           region,
		DefaultKeyName:   keyPairName,
		SSLKeyPair:       sslKeyPair,
	}

	manifest, err := b.manifestBuilder.Build(manifestProperties)
	if err != nil {
		return ssl.KeyPair{}, err
	}

	return manifest.DirectorSSLKeyPair(), nil
}
