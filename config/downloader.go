package config

import (
	"github.com/cloudfoundry/bosh-bootloader/backends"
)

type Downloader struct {
	provider backends.Provider
}

func NewDownloader(provider backends.Provider) Downloader {
	return Downloader{provider: provider}
}

func (d Downloader) DownloadAndPrepareState(flags GlobalFlags) error {
	backend, err := d.provider.Client(flags.IAAS)
	if err != nil {
		return err
	}

	var config backends.Config
	switch flags.IAAS {
	case "aws":
		config = backends.Config{
			Dest:               flags.StateDir,
			Bucket:             flags.StateBucket,
			Region:             flags.AWSRegion,
			AWSAccessKeyID:     flags.AWSAccessKeyID,
			AWSSecretAccessKey: flags.AWSSecretAccessKey,
		}

	case "gcp":
		config = backends.Config{
			Dest:                 flags.StateDir,
			Bucket:               flags.StateBucket,
			Region:               flags.GCPRegion,
			GCPServiceAccountKey: flags.GCPServiceAccountKey,
		}
	}

	return backend.GetState(config, flags.EnvID)
}
