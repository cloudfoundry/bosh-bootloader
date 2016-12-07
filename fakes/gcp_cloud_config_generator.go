package fakes

import "github.com/cloudfoundry/bosh-bootloader/cloudconfig/gcp"

type GCPCloudConfigGenerator struct {
	GenerateCall struct {
		Receives struct {
			CloudConfigInput gcp.CloudConfigInput
		}
		Returns struct {
			CloudConfig gcp.CloudConfig
			Error       error
		}
		CallCount int
	}
}

func (g *GCPCloudConfigGenerator) Generate(cloudConfigInput gcp.CloudConfigInput) (gcp.CloudConfig, error) {
	g.GenerateCall.CallCount++
	g.GenerateCall.Receives.CloudConfigInput = cloudConfigInput

	return g.GenerateCall.Returns.CloudConfig, g.GenerateCall.Returns.Error
}
