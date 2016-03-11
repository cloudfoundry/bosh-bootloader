package unsupported

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
)

type BOSHDeployer struct {
	manifestBuilder boshInitManifestBuilder
}

type boshInitManifestBuilder interface {
	Build(boshinit.ManifestProperties) (boshinit.Manifest, boshinit.ManifestProperties, error)
}

func NewBOSHDeployer(manifestBuilder boshInitManifestBuilder) BOSHDeployer {
	return BOSHDeployer{
		manifestBuilder: manifestBuilder,
	}
}

func (b BOSHDeployer) Deploy(stack cloudformation.Stack, cloudformationClient cloudformation.Client, region string, keyPairName string, sslKeyPair ssl.KeyPair) (ssl.KeyPair, error) {
	_, manifestProperties, err := b.manifestBuilder.Build(boshinit.ManifestProperties{
		SubnetID:         stack.Outputs["BOSHSubnet"],
		AvailabilityZone: stack.Outputs["BOSHSubnetAZ"],
		ElasticIP:        stack.Outputs["BOSHEIP"],
		AccessKeyID:      stack.Outputs["BOSHUserAccessKey"],
		SecretAccessKey:  stack.Outputs["BOSHUserSecretAccessKey"],
		SecurityGroup:    stack.Outputs["BOSHSecurityGroup"],
		Region:           region,
		DefaultKeyName:   keyPairName,
		SSLKeyPair:       sslKeyPair,
	})
	if err != nil {
		return ssl.KeyPair{}, err
	}

	return manifestProperties.SSLKeyPair, nil
}
