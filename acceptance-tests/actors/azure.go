package actors

import (
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
)

type Azure struct {
	groupsClient   *resources.GroupsClient
	subscriptionID string
	tenantID       string
	clientID       string
	clientSecret   string
}

func NewAzure(config acceptance.Config) Azure {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, config.AzureTenantID)
	if err != nil {
		panic(err)
	}

	servicePrincipalToken, err := adal.NewServicePrincipalToken(*oauthConfig, config.AzureClientID, config.AzureClientSecret, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		panic(err)
	}

	gc := resources.NewGroupsClient(config.AzureSubscriptionID)
	gc.ManagementClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	gc.ManagementClient.Sender = autorest.CreateSender(autorest.AsIs())

	return Azure{
		groupsClient:   &gc,
		subscriptionID: config.AzureSubscriptionID,
		tenantID:       config.AzureTenantID,
		clientID:       config.AzureClientID,
		clientSecret:   config.AzureClientSecret,
	}
}

func (a Azure) GetResourceGroup(resourceGroupName string) (bool, error) {
	_, err := a.groupsClient.Get(resourceGroupName)
	if err != nil {
		return false, err
	}

	return true, nil
}
