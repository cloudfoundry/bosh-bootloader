package fakes

type RandomPort struct {
	GetPortCall struct {
		CallCount int
		Returns   struct {
			Port  string
			Error error
		}
	}
}

func (r *RandomPort) GetPort() (string, error) {
	r.GetPortCall.CallCount++

	return r.GetPortCall.Returns.Port, r.GetPortCall.Returns.Error
}
