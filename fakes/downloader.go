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
	d.DownloadCall.CallCount++
	d.DownloadCall.Receives.GlobalFlags = flags

	return d.DownloadCall.Returns.Error
}
