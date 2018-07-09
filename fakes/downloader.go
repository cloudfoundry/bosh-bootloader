package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/config"
)

type Downloader struct {
	DownloadCall struct {
		CallCount int
		Receives  struct {
			GlobalFlags config.GlobalFlags
		}

		Returns struct {
			Error error
		}
	}
}

func (d *Downloader) DownloadAndPrepareState(flags config.GlobalFlags) error {
	d.DownloadAndPrepareStateCall.CallCount++
	d.DownloadAndPrepareStateCall.Receives.GlobalFlags = flags

	return d.DownloadAndPrepareStateCall.Returns.Error
}
