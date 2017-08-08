package azure

type AzureClient struct{}

func NewClient() AzureClient {
	return AzureClient{}
}

func (a AzureClient) ValidateCredentials(tenantID, clientID, clientSecret string) error {
	return nil
}
