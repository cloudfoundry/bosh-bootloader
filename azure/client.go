package azure

import (
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

type AzureClient struct{}

func NewClient() AzureClient {
	return AzureClient{}
}

func (a AzureClient) ValidateCredentials(subscriptionID, tenantID, clientID, clientSecret string) error {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, tenantID)
	if err != nil {
		return err
	}
	_, err = adal.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return err
	}
	gc := resources.NewGroupClient(subscriptionID)
	_, err = gc.CheckExistence("invalid-resource-name", "", "", "", "")
	if err != nil {
		return err
	}

	return nil
}
