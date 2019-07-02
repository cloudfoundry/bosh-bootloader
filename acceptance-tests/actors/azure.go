package actors

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"

	. "github.com/onsi/gomega"
)

type azureLBHelper struct {
	applicationGatewaysClient *network.ApplicationGatewaysClient
	loadBalancersClient       *network.LoadBalancersClient
}

func NewAzureLBHelper(config acceptance.Config) azureLBHelper {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, config.AzureTenantID)
	Expect(err).NotTo(HaveOccurred())

	servicePrincipalToken, err := adal.NewServicePrincipalToken(*oauthConfig, config.AzureClientID, config.AzureClientSecret, azure.PublicCloud.ResourceManagerEndpoint)
	Expect(err).NotTo(HaveOccurred())

	agc := network.NewApplicationGatewaysClient(config.AzureSubscriptionID)
	agc.ManagementClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	agc.ManagementClient.Sender = autorest.CreateSender(autorest.AsIs())

	lbc := network.NewLoadBalancersClient(config.AzureSubscriptionID)
	lbc.ManagementClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	lbc.ManagementClient.Sender = autorest.CreateSender(autorest.AsIs())

	return azureLBHelper{
		loadBalancersClient:       &lbc,
		applicationGatewaysClient: &agc,
	}
}

func (z azureLBHelper) getLoadBalancer(resourceGroupName, loadBalancerName string) (bool, error) {
	_, err := z.loadBalancersClient.Get(fmt.Sprintf("%s-bosh", resourceGroupName), loadBalancerName, "")
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

func (z azureLBHelper) getApplicationGateway(resourceGroupName, applicationGatewayName string) (bool, error) {
	_, err := z.applicationGatewaysClient.Get(fmt.Sprintf("%s-bosh", resourceGroupName), applicationGatewayName)
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

func (z azureLBHelper) GetLBArgs() []string {
	return []string{"--lb-type", "concourse"}
}

func (z azureLBHelper) VerifyCloudConfigExtensions(vmExtensions []string) {
	Expect(vmExtensions).To(ContainElement("lb"))
}

func (z azureLBHelper) ConfirmLBsExist(envID string) {
	exists, err := z.getLoadBalancer(envID, fmt.Sprintf("%s-concourse-lb", envID))
	Expect(err).NotTo(HaveOccurred())
	Expect(exists).To(BeTrue())
}

func (z azureLBHelper) ConfirmNoLBsExist(envID string) {
	exists, err := z.getLoadBalancer(envID, fmt.Sprintf("%s-concourse-lb", envID))
	Expect(err).NotTo(HaveOccurred())
	Expect(exists).To(BeFalse())
}

func (z azureLBHelper) VerifyBblLBOutput(stdout string) {
	Expect(stdout).To(MatchRegexp("Concourse LB:.*"))
}

func (z azureLBHelper) ConfirmNoStemcellsExist(stemcellIDs []string) {}
