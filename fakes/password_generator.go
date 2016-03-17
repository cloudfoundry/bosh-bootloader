package fakes

type PasswordGenerator struct {
	GenerateCall struct {
		Returns struct {
			Password string
			Error    error
		}
	}
}

func (p *PasswordGenerator) Generate() (string, error) {
	return p.GenerateCall.Returns.Password, p.GenerateCall.Returns.Error
}
