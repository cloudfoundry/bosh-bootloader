package fakes

type GuidGenerator struct {
	GenerateCall struct {
		Receives struct {
			CallCount int
		}
		Returns struct {
			Output string
			Error  error
		}
	}
}

func (g *GuidGenerator) Generate() (string, error) {
	g.GenerateCall.Receives.CallCount++
	return g.GenerateCall.Returns.Output, g.GenerateCall.Returns.Error
}
