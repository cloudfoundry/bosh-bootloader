package fakes

type PatchDetector struct {
	FindCall struct {
		CallCount int
		Returns   struct {
			Error error
		}
	}
}

func (p *PatchDetector) Find() error {
	p.FindCall.CallCount++

	return p.FindCall.Returns.Error
}
