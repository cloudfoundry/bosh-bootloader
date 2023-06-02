package actors

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"                         //nolint:staticcheck
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork" //nolint:staticcheck
	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"

	. "github.com/onsi/gomega"
)

type azureLBHelper struct {
	applicationGatewaysClient *armnetwork.ApplicationGatewaysClient
	loadBalancersClient       *armnetwork.LoadBalancersClient
}

func NewAzureLBHelper(config acceptance.Config) azureLBHelper {
	credential, err := azidentity.NewClientSecretCredential(config.AzureTenantID, config.AzureClientID, config.AzureClientSecret, nil)
	Expect(err).NotTo(HaveOccurred())

	agc, err := armnetwork.NewApplicationGatewaysClient(config.AzureTenantID, credential, nil)
	Expect(err).NotTo(HaveOccurred())

	lbc, err := armnetwork.NewLoadBalancersClient(config.AzureTenantID, credential, nil)
	Expect(err).NotTo(HaveOccurred())

	return azureLBHelper{
		loadBalancersClient:       lbc,
		applicationGatewaysClient: agc,
	}
}

func (z azureLBHelper) getLoadBalancer(resourceGroupName, loadBalancerName string) (bool, error) {
	_, err := z.loadBalancersClient.Get(context.TODO(), fmt.Sprintf("%s-bosh", resourceGroupName), loadBalancerName, nil)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (z azureLBHelper) getApplicationGateway(resourceGroupName, applicationGatewayName string) (bool, error) {
	_, err := z.applicationGatewaysClient.Get(context.TODO(), fmt.Sprintf("%s-bosh", resourceGroupName), applicationGatewayName, nil)
	if err != nil {
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
