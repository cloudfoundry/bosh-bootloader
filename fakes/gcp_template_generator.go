package fakes

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
			Region string
			LBType string
			Domain string
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

func (g *GCPTemplateGenerator) Generate(region string, lbType string, domain string) string {
	g.GenerateCall.CallCount++
	g.GenerateCall.Receives.Region = region
	g.GenerateCall.Receives.LBType = lbType
	g.GenerateCall.Receives.Domain = domain
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
