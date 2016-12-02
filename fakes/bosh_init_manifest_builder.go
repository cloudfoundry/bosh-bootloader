package fakes

import "github.com/cloudfoundry/bosh-bootloader/boshinit/manifests"

type BOSHInitManifestBuilder struct {
	BuildCall struct {
		Receives struct {
			Properties manifests.ManifestProperties
			IAAS       string
		}
		Returns struct {
			Manifest   manifests.Manifest
			Properties manifests.ManifestProperties
			Error      error
		}
	}
}

func (b *BOSHInitManifestBuilder) Build(iaas string, properties manifests.ManifestProperties) (manifests.Manifest, manifests.ManifestProperties, error) {
	b.BuildCall.Receives.Properties = properties
	b.BuildCall.Receives.IAAS = iaas

	return b.BuildCall.Returns.Manifest, b.BuildCall.Returns.Properties, b.BuildCall.Returns.Error
}
