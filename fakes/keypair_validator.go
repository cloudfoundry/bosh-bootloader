package fakes

type KeyPairValidator struct {
	ValidateCall struct {
		CallCount int
		Receives  struct {
			PEMData []byte
		}
		Returns struct {
			Error error
		}
	}
}

func (v *KeyPairValidator) Validate(pemData []byte) error {
	v.ValidateCall.CallCount++
	v.ValidateCall.Receives.PEMData = pemData

	return v.ValidateCall.Returns.Error
}
