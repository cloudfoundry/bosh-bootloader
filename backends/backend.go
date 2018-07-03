package backends

import (
	"context"
	"io"

	"github.com/lytics/cloudstorage"
)

type Backend interface {
	GetState(cloudstorage.Config, string) (io.ReadCloser, error)
}

type backend struct{}

func (b backend) GetState(config cloudstorage.Config, name string) (io.ReadCloser, error) {
	store, err := cloudstorage.NewStore(&config)
	if err != nil {
		return nil, err
	}

	tarball, err := store.Get(context.Background(), name)
	if err != nil {
		return nil, err
	}

	return tarball.Open(cloudstorage.ReadOnly)
}
