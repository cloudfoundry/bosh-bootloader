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

func (a *AWSCredentialValidator) Validate(accessKeyID string, secretAccessKey string, region string) error {
	a.ValidateCall.CallCount++
	a.ValidateCall.Receives.AccessKeyID = accessKeyID
	a.ValidateCall.Receives.SecretAccessKey = secretAccessKey
	a.ValidateCall.Receives.Region = region
	return a.ValidateCall.Returns.Error
}
