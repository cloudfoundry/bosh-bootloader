package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type GCPTemplateGenerator struct {
	GenerateBackendServiceCall struct {
		CallCount int
		Receives  struct {
			Region string
		}
		Returns struct {
			Template string
		}
	}
	GenerateCall struct {
		CallCount int
		Receives  struct {
			State storage.State
		}
		Returns struct {
			Template string
		}
	}
	GenerateInstanceGroupsCall struct {
		CallCount int
		Receives  struct {
			Region string
		}
		Returns struct {
			Template string
		}
	}
}

func (g *GCPTemplateGenerator) Generate(state storage.State) string {
	g.GenerateCall.CallCount++
	g.GenerateCall.Receives.State = state
	return g.GenerateCall.Returns.Template
}

func (g *GCPTemplateGenerator) GenerateBackendService(region string) string {
	g.GenerateBackendServiceCall.CallCount++
	g.GenerateBackendServiceCall.Receives.Region = region
	return g.GenerateBackendServiceCall.Returns.Template
}

func (g *GCPTemplateGenerator) GenerateInstanceGroups(region string) string {
	g.GenerateInstanceGroupsCall.CallCount++
	g.GenerateInstanceGroupsCall.Receives.Region = region
	return g.GenerateInstanceGroupsCall.Returns.Template
}
