package azure

import (
	"context"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"

	"golang.org/x/oauth2/jwt"
)

func azureHTTPClientFunc(config *jwt.Config) *http.Client {
	return config.Client(context.Background())
}

var azureHTTPClient = azureHTTPClientFunc

type ClientProvider struct {
	client Client
}

func NewClientProvider() *ClientProvider {
	return &ClientProvider{}
}

func (p *ClientProvider) SetConfig(subscriptionID, tenantID, clientID, clientSecret string) error {
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

	vmsClient := compute.NewVirtualMachinesClient(subscriptionID)
	vmsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	vmsClient.Sender = autorest.CreateSender(autorest.AsIs())

	p.client = Client{
		azureVMsClient: vmsClient,
	}

	_, err = ac.List()
	if err != nil {
		return err
	}

	return nil
}

func (p *ClientProvider) Client() Client {
	return p.client
}
