package fakes

type CredentialValidator struct {
	ValidateAWSCall struct {
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
	ValidateGCPCall struct {
		CallCount int
		Returns   struct {
			Error error
		}
		Receives struct {
			ProjectID         string
			ServiceAccountKey string
			Region            string
			Zone              string
		}
	}
}

func (c *CredentialValidator) ValidateAWS() error {
	c.ValidateAWSCall.CallCount++
	return c.ValidateAWSCall.Returns.Error
}

func (c *CredentialValidator) ValidateGCP() error {
	c.ValidateGCPCall.CallCount++
	return c.ValidateGCPCall.Returns.Error
}
