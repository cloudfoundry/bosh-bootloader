package fakes

import (
	"io"

	"github.com/cloudfoundry/bosh-bootloader/config"
)

type Downloader struct {
	DownloadCall struct {
		CallCount int
		Receives  struct {
			GlobalFlags config.GlobalFlags
		}

		Returns struct {
			State io.ReadCloser
			Error error
		}
	}
}

func (d *Downloader) Download(flags config.GlobalFlags) (io.ReadCloser, error) {
	d.DownloadCall.CallCount++
	d.DownloadCall.Receives.GlobalFlags = flags

	return d.DownloadCall.Returns.State, d.DownloadCall.Returns.Error
}
