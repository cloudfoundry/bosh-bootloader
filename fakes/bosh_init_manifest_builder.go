package fakes

import "github.com/cloudfoundry/bosh-bootloader/boshinit/manifests"

type BOSHInitManifestBuilder struct {
	BuildCall struct {
		Receives struct {
			Properties manifests.ManifestProperties
		}
		Returns struct {
			Manifest   manifests.Manifest
			Properties manifests.ManifestProperties
			Error      error
		}
	}
}

func (b *BOSHInitManifestBuilder) Build(properties manifests.ManifestProperties) (manifests.Manifest, manifests.ManifestProperties, error) {
	b.BuildCall.Receives.Properties = properties

	return b.BuildCall.Returns.Manifest, b.BuildCall.Returns.Properties, b.BuildCall.Returns.Error
}
