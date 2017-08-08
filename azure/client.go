package azure

import (
	"github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/go-autorest/autorest"
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
	servicePrincipalToken, err := adal.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return err
	}

	ac := storage.NewAccountsClient(subscriptionID)
	ac.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	ac.Sender = autorest.CreateSender(autorest.AsIs())

	_, err = ac.List()
	if err != nil {
		return err
	}

	return nil
}
