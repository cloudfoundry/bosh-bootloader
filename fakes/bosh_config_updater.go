package fakes

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type BOSHConfigUpdater struct {
	InitializeAuthenticatedCLICall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			AuthenticatedCLIRunner bosh.AuthenticatedCLIRunner
			Error                  error
		}
	}
	UpdateCloudConfigCall struct {
		CallCount int
		Receives  struct {
			AuthenticatedCLIRunner bosh.AuthenticatedCLIRunner
			Filepath               string
			OpsFilepaths           []string
			VarsFilepath           string
		}
		Returns struct {
			Error error
		}
	}
	UpdateRuntimeConfigCall struct {
		CallCount int
		Receives  struct {
			AuthenticatedCLIRunner bosh.AuthenticatedCLIRunner
			Filepath               string
			OpsFilepaths           []string
			Name                   string
		}
		Returns struct {
			Error error
		}
	}
}

func (c *BOSHConfigUpdater) InitializeAuthenticatedCLI(state storage.State) (bosh.AuthenticatedCLIRunner, error) {
	c.InitializeAuthenticatedCLICall.CallCount++

	c.InitializeAuthenticatedCLICall.Receives.State = state

	return c.InitializeAuthenticatedCLICall.Returns.AuthenticatedCLIRunner, c.InitializeAuthenticatedCLICall.Returns.Error
}

func (c *BOSHConfigUpdater) UpdateRuntimeConfig(authenticatedCLIRunner bosh.AuthenticatedCLIRunner, filepath string, opsFilepaths []string, name string) error {
	c.UpdateRuntimeConfigCall.CallCount++

	c.UpdateRuntimeConfigCall.Receives.AuthenticatedCLIRunner = authenticatedCLIRunner
	c.UpdateRuntimeConfigCall.Receives.Filepath = filepath
	c.UpdateRuntimeConfigCall.Receives.OpsFilepaths = opsFilepaths
	c.UpdateRuntimeConfigCall.Receives.Name = name

	return c.UpdateRuntimeConfigCall.Returns.Error
}

func (c *BOSHConfigUpdater) UpdateCloudConfig(authenticatedCLIRunner bosh.AuthenticatedCLIRunner, filepath string, opsFilepaths []string, varsFilepath string) error {
	c.UpdateCloudConfigCall.CallCount++

	c.UpdateCloudConfigCall.Receives.AuthenticatedCLIRunner = authenticatedCLIRunner
	c.UpdateCloudConfigCall.Receives.Filepath = filepath
	c.UpdateCloudConfigCall.Receives.OpsFilepaths = opsFilepaths
	c.UpdateCloudConfigCall.Receives.VarsFilepath = varsFilepath

	return c.UpdateCloudConfigCall.Returns.Error
}
