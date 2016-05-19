package fakes

type AWSCredentialValidator struct {
	ValidateCall struct {
		CallCount int
		Returns   struct {
			Error error
		}
		Receives struct {
			AccessKeyID     string
			SecretAccessKey string
			Region          string
		}
	}
}

func (a *AWSCredentialValidator) Validate() error {
	a.ValidateCall.CallCount++
	return a.ValidateCall.Returns.Error
}
