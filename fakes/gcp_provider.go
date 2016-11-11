package fakes

import "github.com/cloudfoundry/bosh-bootloader/gcp"

type GCPProvider struct {
	GetServiceCall struct {
		CallCount int
		Returns   struct {
			Service gcp.ServiceWrapper
		}
	}
	SetConfigCall struct {
		CallCount int
		Receives  struct {
			ServiceAccountKey string
		}
		Returns struct {
			Error error
		}
	}
}

func (g *GCPProvider) GetService() gcp.ServiceWrapper {
	g.GetServiceCall.CallCount++

	return g.GetServiceCall.Returns.Service
}

func (g *GCPProvider) SetConfig(serviceAccountKey string) error {
	g.SetConfigCall.CallCount++
	g.SetConfigCall.Receives.ServiceAccountKey = serviceAccountKey

	return g.SetConfigCall.Returns.Error
}
