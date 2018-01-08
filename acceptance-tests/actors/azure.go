package actors

import (
	"encoding/base64"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"

	. "github.com/onsi/gomega"
)

type Azure struct {
	groupsClient              *resources.GroupsClient
	virtualMachinesClient     *compute.VirtualMachinesClient
	applicationGatewaysClient *network.ApplicationGatewaysClient
	subscriptionID            string
	tenantID                  string
	clientID                  string
	clientSecret              string
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

	vmc := compute.NewVirtualMachinesClient(config.AzureSubscriptionID)
	vmc.ManagementClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	vmc.ManagementClient.Sender = autorest.CreateSender(autorest.AsIs())

	agc := network.NewApplicationGatewaysClient(config.AzureSubscriptionID)
	agc.ManagementClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	agc.ManagementClient.Sender = autorest.CreateSender(autorest.AsIs())

	return Azure{
		groupsClient:              &gc,
		virtualMachinesClient:     &vmc,
		applicationGatewaysClient: &agc,
		subscriptionID:            config.AzureSubscriptionID,
		tenantID:                  config.AzureTenantID,
		clientID:                  config.AzureClientID,
		clientSecret:              config.AzureClientSecret,
	}
}

func (a Azure) GetApplicationGateway(resourceGroupName, applicationGatewayName string) (bool, error) {
	_, err := a.applicationGatewaysClient.Get(fmt.Sprintf("%s-bosh", resourceGroupName), applicationGatewayName)
	if err != nil {
		if aerr, ok := err.(autorest.DetailedError); ok {
			if aerr.StatusCode.(int) == 404 {
				return false, nil
			}
		}
		return false, err
	}

	return true, nil
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

type azureIaasLbHelper struct {
	azure Azure
}

func (z azureIaasLbHelper) GetLBArgs() []string {
	pfx_data, err := base64.StdEncoding.DecodeString(testhelpers.PFX_BASE64)
	Expect(err).NotTo(HaveOccurred())

	certPath, err := testhelpers.WriteByteContentsToTempFile(pfx_data)
	Expect(err).NotTo(HaveOccurred())

	keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.PFX_PASSWORD)
	Expect(err).NotTo(HaveOccurred())
	return []string{
		"--lb-type", "cf",
		"--lb-cert", certPath,
		"--lb-key", keyPath,
	}
}

func (z azureIaasLbHelper) ConfirmLBsExist(envID string) {
	exists, err := z.azure.GetApplicationGateway(envID, fmt.Sprintf("%s-app-gateway", envID))
	Expect(err).NotTo(HaveOccurred())
	Expect(exists).To(BeTrue())
}

func (z azureIaasLbHelper) ConfirmNoLBsExist(envID string) {
	exists, err := z.azure.GetApplicationGateway(envID, fmt.Sprintf("%s-app-gateway", envID))
	Expect(err).NotTo(HaveOccurred())
	Expect(exists).To(BeFalse())
}
