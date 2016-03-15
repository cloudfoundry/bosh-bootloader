package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"

type CloudProviderManifestBuilder struct {
	BuildCall struct {
		Returns struct {
			Error error
		}
	}
}

func (c CloudProviderManifestBuilder) Build(manifestProperties boshinit.ManifestProperties) (boshinit.CloudProvider, boshinit.ManifestProperties, error) {
	return boshinit.CloudProvider{}, manifestProperties, c.BuildCall.Returns.Error
}
