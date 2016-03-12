package fakes

type Executable struct {
	RunCall struct {
		CallCount int
		Stub      func() error
		Returns   struct {
			Error error
		}
	}
}

func (e *Executable) Run() error {
	e.RunCall.CallCount++

	if e.RunCall.Stub != nil {
		return e.RunCall.Stub()
	}

	return e.RunCall.Returns.Error
}
