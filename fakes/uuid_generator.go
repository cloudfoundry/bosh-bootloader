package fakes

type GenerateReturn struct {
	String string
	Error  error
}

type UUIDGenerator struct {
	GenerateCall struct {
		Returns   []GenerateReturn
		CallCount int
	}
}

func (u *UUIDGenerator) Generate() (string, error) {
	defer func() { u.GenerateCall.CallCount += 1 }()
	return u.GenerateCall.Returns[u.GenerateCall.CallCount].String, u.GenerateCall.Returns[u.GenerateCall.CallCount].Error
}
