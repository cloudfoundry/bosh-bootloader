package fakes

type AzureClient struct {
	ValidateCredentialsCall struct {
		CallCount int
		Receives  struct {
			TenantID     string
			ClientID     string
			ClientSecret string
		}
		Returns struct {
			Error error
		}
	}
}

func (a *AzureClient) ValidateCredentials(tenantID, clientID, clientSecret string) error {
	a.ValidateCredentialsCall.CallCount++
	a.ValidateCredentialsCall.Receives.TenantID = tenantID
	a.ValidateCredentialsCall.Receives.ClientID = clientID
	a.ValidateCredentialsCall.Receives.ClientSecret = clientSecret
	return a.ValidateCredentialsCall.Returns.Error
}
