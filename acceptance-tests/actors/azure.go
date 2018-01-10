package actors

import (
	"encoding/base64"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"

	. "github.com/onsi/gomega"
)

type azureLBHelper struct {
	applicationGatewaysClient *network.ApplicationGatewaysClient
}

func NewAzureLBHelper(config acceptance.Config) azureLBHelper {
	oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, config.AzureTenantID)
	if err != nil {
		panic(err)
	}

	servicePrincipalToken, err := adal.NewServicePrincipalToken(*oauthConfig, config.AzureClientID, config.AzureClientSecret, azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		panic(err)
	}

	agc := network.NewApplicationGatewaysClient(config.AzureSubscriptionID)
	agc.ManagementClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	agc.ManagementClient.Sender = autorest.CreateSender(autorest.AsIs())

	return azureLBHelper{
		applicationGatewaysClient: &agc,
	}
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

func (z azureLBHelper) ConfirmLBsExist(envID string) {
	exists, err := z.getApplicationGateway(envID, fmt.Sprintf("%s-app-gateway", envID))
	Expect(err).NotTo(HaveOccurred())
	Expect(exists).To(BeTrue())
}

func (z azureLBHelper) ConfirmNoLBsExist(envID string) {
	exists, err := z.getApplicationGateway(envID, fmt.Sprintf("%s-app-gateway", envID))
	Expect(err).NotTo(HaveOccurred())
	Expect(exists).To(BeFalse())
}

func (z azureLBHelper) VerifyBblLBOutput(stdout string) {
	Expect(stdout).To(MatchRegexp("CF LB:.*"))
}
