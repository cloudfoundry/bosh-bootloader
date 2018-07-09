package backends

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/araddon/gou"
	"github.com/cloudfoundry/bbl-state-resource/storage"
	"github.com/lytics/cloudstorage"
	"github.com/lytics/cloudstorage/awss3"
	"github.com/mholt/archiver"
)

type Config struct {
	AWSAccessKeyID       string
	AWSSecretAccessKey   string
	GCPServiceAccountKey string
	Bucket               string
	Region               string
	Dest                 string
}

type Provider interface {
	Client(string) (Backend, error)
}

func NewProvider() Provider {
	return provider{}
}

type provider struct{}

func (p provider) Client(iaas string) (Backend, error) {
	switch iaas {
	case "aws":
		return cloudStorageBackend{}, nil
	case "gcp":
		return gcsStateBackend{}, nil
	default:
		return nil, fmt.Errorf("remote state storage is unsupported for %s environments", iaas)
	}
}

type Backend interface {
	GetState(Config, string) error
}

type cloudStorageBackend struct{}

func (c cloudStorageBackend) GetState(config Config, name string) error {
	awsAuthSettings := make(gou.JsonHelper)
	awsAuthSettings[awss3.ConfKeyAccessKey] = config.AWSAccessKeyID
	awsAuthSettings[awss3.ConfKeyAccessSecret] = config.AWSSecretAccessKey

	csConfig := cloudstorage.Config{
		Type:       awss3.StoreType,
		AuthMethod: awss3.AuthAccessKey,
		Bucket:     config.Bucket,
		Settings:   awsAuthSettings,
		Region:     config.Region,
	}

	store, err := cloudstorage.NewStore(&csConfig)
	if err != nil {
		return err
	}

	tarball, err := store.Get(context.Background(), name)
	if err != nil {
		return err
	}

	stateTar, err := tarball.Open(cloudstorage.ReadOnly)
	if err != nil {
		return err
	}

	err = archiver.TarGz.Read(stateTar, config.Dest)
	if err != nil {
		return fmt.Errorf("unable to untar state dir: %s", err)
	}

	return nil
}

type gcsStateBackend struct{}

func (g gcsStateBackend) GetState(config Config, name string) error {
	key, err := g.getGCPServiceAccountKey(config.GCPServiceAccountKey)
	if err != nil {
		return fmt.Errorf("could not read GCP service account key: %s", err)
	}

	gcsClient, err := storage.NewStorageClient(key, name, config.Bucket)
	if err != nil {
		return fmt.Errorf("could not create GCS client: %s", err)
	}

	_, err = gcsClient.Download(config.Dest)
	if err != nil {
		return fmt.Errorf("downloading remote state from GCS: %s", err)
	}

	return nil
}

func (g gcsStateBackend) getGCPServiceAccountKey(key string) (string, error) {
	if _, err := os.Stat(key); err != nil {
		return key, nil
	}

	keyBytes, err := ioutil.ReadFile(key)
	if err != nil {
		return "", fmt.Errorf("Reading key: %v", err)
	}

	return string(keyBytes), nil
}
