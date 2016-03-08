package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"

type BoshInitManifestBuilder struct {
	BuildCall struct {
		Returns struct {
			Manifest boshinit.Manifest
		}
	}
}

func (b BoshInitManifestBuilder) Build() boshinit.Manifest {
	return b.BuildCall.Returns.Manifest
}
