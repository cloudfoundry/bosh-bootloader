package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"

type JobsManifestBuilder struct {
	BuildCall struct {
		Returns struct {
			Error error
		}
	}
}

func (j JobsManifestBuilder) Build(manifestProperties boshinit.ManifestProperties) ([]boshinit.Job, boshinit.ManifestProperties, error) {
	return nil, manifestProperties, j.BuildCall.Returns.Error
}
