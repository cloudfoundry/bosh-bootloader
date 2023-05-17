package azure

import (
	"github.com/Azure/azure-sdk-for-go/arm/compute"              //nolint:staticcheck
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"  //nolint:staticcheck
	azurestorage "github.com/Azure/azure-sdk-for-go/arm/storage" //nolint:staticcheck
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

func NewClient(azureConfig storage.Azure) (Client, error) {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, azureConfig.TenantID)
	if err != nil {
		return Client{}, err
	}

	servicePrincipalToken, err := adal.NewServicePrincipalToken(*oauthConfig, azureConfig.ClientID, azureConfig.ClientSecret, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		return Client{}, err
	}

	ac := azurestorage.NewAccountsClient(azureConfig.SubscriptionID)
	ac.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	ac.Sender = autorest.CreateSender(autorest.AsIs())

	vmsClient := compute.NewVirtualMachinesClient(azureConfig.SubscriptionID)
	vmsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	vmsClient.Sender = autorest.CreateSender(autorest.AsIs())

	groupsClient := resources.NewGroupsClient(azureConfig.SubscriptionID)
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	groupsClient.Sender = autorest.CreateSender(autorest.AsIs())

	client := Client{
		azureVMsClient:    vmsClient,
		azureGroupsClient: groupsClient,
	}

	_, err = ac.List()
	if err != nil {
		return Client{}, err
	}

	return client, nil
}
