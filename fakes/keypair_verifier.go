package fakes

type KeyPairVerifier struct {
	VerifyCall struct {
		Receives struct {
			Fingerprint string
			PEMData     []byte
		}
		Returns struct {
			Error error
		}
	}
}

func (v *KeyPairVerifier) Verify(fingerprint string, pemData []byte) error {
	v.VerifyCall.Receives.Fingerprint = fingerprint
	v.VerifyCall.Receives.PEMData = pemData

	return v.VerifyCall.Returns.Error
}
