package fakes

import (
	"io"

	"github.com/lytics/cloudstorage"
)

type Backend struct {
	GetStateCall struct {
		CallCount int
		Receives  struct {
			Config cloudstorage.Config
			Name   string
		}

		Returns struct {
			State io.ReadCloser
			Error error
		}
	}
}

func (b *Backend) GetState(config cloudstorage.Config, name string) (io.ReadCloser, error) {
	b.GetStateCall.CallCount++
	b.GetStateCall.Receives.Config = config
	b.GetStateCall.Receives.Name = name

	return b.GetStateCall.Returns.State, b.GetStateCall.Returns.Error
}
