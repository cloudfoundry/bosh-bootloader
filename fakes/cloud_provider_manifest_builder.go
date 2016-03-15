package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"

type CloudProviderManifestBuilder struct {
	BuildCall struct {
		Returns struct {
			Error error
		}
	}
}

func (c CloudProviderManifestBuilder) Build(manifestProperties boshinit.ManifestProperties) (boshinit.CloudProvider, error) {
	return boshinit.CloudProvider{}, c.BuildCall.Returns.Error
}
