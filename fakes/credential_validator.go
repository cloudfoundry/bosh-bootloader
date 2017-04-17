package fakes

type CredentialValidator struct {
	ValidateCall struct {
		CallCount int
		Returns   struct {
			Error error
		}
	}
}

func (c *CredentialValidator) Validate() error {
	c.ValidateCall.CallCount++
	return c.ValidateCall.Returns.Error
}
