package fakes

import "github.com/cloudfoundry/bosh-bootloader/bosh"

type BOSHClient struct {
	UpdateCloudConfigCall struct {
		CallCount int
		Receives  struct {
			Yaml []byte
		}
		Returns struct {
			Error error
		}
	}

	InfoCall struct {
		CallCount int
		Returns   struct {
			Info  bosh.Info
			Error error
		}
	}
}

func (c *BOSHClient) UpdateCloudConfig(yaml []byte) error {
	c.UpdateCloudConfigCall.CallCount++
	c.UpdateCloudConfigCall.Receives.Yaml = yaml
	return c.UpdateCloudConfigCall.Returns.Error
}

func (c *BOSHClient) Info() (bosh.Info, error) {
	c.InfoCall.CallCount++
	return c.InfoCall.Returns.Info, c.InfoCall.Returns.Error
}
