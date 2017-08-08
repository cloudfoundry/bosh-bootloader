package fakes

type AzureClient struct {
	ValidateCredentialsCall struct {
		CallCount int
		Receives  struct {
			SubscriptionID string
			TenantID       string
			ClientID       string
			ClientSecret   string
		}
		Returns struct {
			Error error
		}
	}
}

func (a *AzureClient) ValidateCredentials(subscriptionID, tenantID, clientID, clientSecret string) error {
	a.ValidateCredentialsCall.CallCount++
	a.ValidateCredentialsCall.Receives.SubscriptionID = subscriptionID
	a.ValidateCredentialsCall.Receives.TenantID = tenantID
	a.ValidateCredentialsCall.Receives.ClientID = clientID
	a.ValidateCredentialsCall.Receives.ClientSecret = clientSecret
	return a.ValidateCredentialsCall.Returns.Error
}
