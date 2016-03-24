package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/boshinit/manifests"

type CloudProviderManifestBuilder struct {
	BuildCall struct {
		Returns struct {
			Error error
		}
	}
}

func (c CloudProviderManifestBuilder) Build(manifestProperties manifests.ManifestProperties) (manifests.CloudProvider, manifests.ManifestProperties, error) {
	return manifests.CloudProvider{}, manifestProperties, c.BuildCall.Returns.Error
}
