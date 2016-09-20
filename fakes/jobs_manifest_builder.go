package fakes

import "github.com/cloudfoundry/bosh-bootloader/boshinit/manifests"

type JobsManifestBuilder struct {
	BuildCall struct {
		Returns struct {
			Error error
		}
	}
}

func (j JobsManifestBuilder) Build(manifestProperties manifests.ManifestProperties) ([]manifests.Job, manifests.ManifestProperties, error) {
	return nil, manifestProperties, j.BuildCall.Returns.Error
}
