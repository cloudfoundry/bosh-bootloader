package fakes

type GarbageCollector struct {
	RemoveCall struct {
		CallCount int
		Receives  struct {
			Directory string
		}
		Returns struct {
			Error error
		}
	}
}

func (g *GarbageCollector) Remove(directory string) error {
	g.RemoveCall.CallCount++
	g.RemoveCall.Receives.Directory = directory

	return g.RemoveCall.Returns.Error
}
