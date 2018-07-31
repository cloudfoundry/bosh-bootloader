package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"golang.org/x/net/proxy"
)

type BOSHClient struct {
	UpdateConfigCall struct {
		CallCount int
		Receives  struct {
			Type string
			Yaml []byte
		}
		Returns struct {
			Error error
		}
	}

	UpdateCloudConfigCall struct {
		CallCount int
		Receives  struct {
			Yaml []byte
		}
		Returns struct {
			Error error
		}
	}

	ConfigureHTTPClientCall struct {
		CallCount int
		Receives  struct {
			Dialer proxy.Dialer
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

func (c *BOSHClient) UpdateConfig(t string, yaml []byte) error {
	c.UpdateConfigCall.CallCount++
	c.UpdateConfigCall.Receives.Type = t
	c.UpdateConfigCall.Receives.Yaml = yaml
	return c.UpdateConfigCall.Returns.Error
}

func (c *BOSHClient) UpdateCloudConfig(yaml []byte) error {
	c.UpdateCloudConfigCall.CallCount++
	c.UpdateCloudConfigCall.Receives.Yaml = yaml
	return c.UpdateCloudConfigCall.Returns.Error
}

func (c *BOSHClient) ConfigureHTTPClient(dialer proxy.Dialer) {
	c.ConfigureHTTPClientCall.CallCount++
	c.ConfigureHTTPClientCall.Receives.Dialer = dialer
}

func (c *BOSHClient) Info() (bosh.Info, error) {
	c.InfoCall.CallCount++
	return c.InfoCall.Returns.Info, c.InfoCall.Returns.Error
}
