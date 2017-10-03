package actors

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
)

type Azure struct {
	groupsClient          *resources.GroupsClient
	virtualMachinesClient *compute.VirtualMachinesClient
	subscriptionID        string
	tenantID              string
	clientID              string
	clientSecret          string
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

	// blindly copied from groupsClient above
	vmc := compute.NewVirtualMachinesClient(config.AzureSubscriptionID)
	vmc.ManagementClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	vmc.ManagementClient.Sender = autorest.CreateSender(autorest.AsIs())

	return Azure{
		groupsClient:          &gc,
		virtualMachinesClient: &vmc,
		subscriptionID:        config.AzureSubscriptionID,
		tenantID:              config.AzureTenantID,
		clientID:              config.AzureClientID,
		clientSecret:          config.AzureClientSecret,
	}
}

func (a Azure) GetResourceGroup(resourceGroupName string) (bool, error) {
	_, err := a.groupsClient.Get(resourceGroupName)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (a Azure) NetworkHasBOSHDirector(envID string) bool {
	resourceGroupName := fmt.Sprintf("%s-bosh", envID)
	result, err := a.virtualMachinesClient.List(resourceGroupName)
	if err != nil {
		panic(err)
	}

	for _, vm := range *result.Value {
		if *(*vm.Tags)["deployment"] == "bosh" {
			return true
		}
	}

	return false
}
