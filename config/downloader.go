package config

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/araddon/gou"
	"github.com/cloudfoundry/bosh-bootloader/backends"
	"github.com/lytics/cloudstorage"
	"github.com/lytics/cloudstorage/awss3"
	"github.com/lytics/cloudstorage/google"
)

type Downloader struct {
	backend backends.Backend
}

func NewDownloader(backend backends.Backend) Downloader {
	return Downloader{backend: backend}
}

func (d Downloader) Download(flags GlobalFlags) (io.ReadCloser, error) {
	var config cloudstorage.Config

	switch flags.IAAS {
	case "gcp":
		serviceAccountKey := cloudstorage.JwtConf{}

		err := json.Unmarshal([]byte(flags.GCPServiceAccountKey), &serviceAccountKey)
		if err != nil {
			return nil, errors.New("invalid GCP service account key")
		}

		config = cloudstorage.Config{
			Type:       google.StoreType,
			AuthMethod: google.AuthJWTKeySource,
			Project:    serviceAccountKey.ProjectID,
			Bucket:     flags.StateBucket,
			JwtConf:    &serviceAccountKey,
		}
	case "aws":
		awsAuthSettings := make(gou.JsonHelper)
		awsAuthSettings[awss3.ConfKeyAccessKey] = flags.AWSAccessKeyID
		awsAuthSettings[awss3.ConfKeyAccessSecret] = flags.AWSSecretAccessKey

		config = cloudstorage.Config{
			Type:       awss3.StoreType,
			AuthMethod: awss3.AuthAccessKey,
			Bucket:     flags.StateBucket,
			Settings:   awsAuthSettings,
			Region:     flags.AWSRegion,
		}
	}

	return d.backend.GetState(config, flags.EnvID)
}
