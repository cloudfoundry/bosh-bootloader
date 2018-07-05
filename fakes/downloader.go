package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/config"
)

type Downloader struct {
	DownloadAndPrepareStateCall struct {
		CallCount int
		Receives  struct {
			GlobalFlags config.GlobalFlags
		}

		Returns struct {
			StateDirPath string
			Error        error
		}
	}
}

func (d *Downloader) DownloadAndPrepareState(flags config.GlobalFlags) (string, error) {
	d.DownloadAndPrepareStateCall.CallCount++
	d.DownloadAndPrepareStateCall.Receives.GlobalFlags = flags

	return d.DownloadAndPrepareStateCall.Returns.StateDirPath, d.DownloadAndPrepareStateCall.Returns.Error
}
