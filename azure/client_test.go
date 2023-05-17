package azure_test

import (
	"errors"

	"github.com/Azure/azure-sdk-for-go/arm/compute" //nolint:staticcheck
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/mocks"
	"github.com/cloudfoundry/bosh-bootloader/azure"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	Describe("CheckExists", func() {
		var (
			azureClient *fakes.AzureGroupsClient
			client      azure.Client
		)

		BeforeEach(func() {
			azureClient = &fakes.AzureGroupsClient{}
			client = azure.NewClientWithInjectedGroupsClient(azureClient)

			azureClient.CheckExistenceCall.Returns.Response = autorest.Response{
				Response: mocks.NewResponseWithStatus("some-message", 404),
			}
		})

		Context("when the resource group does not exist", func() {
			It("returns false", func() {
				exists, err := client.CheckExists("some-environment")
				Expect(err).NotTo(HaveOccurred())

				Expect(exists).To(BeFalse())
			})
		})

		Context("when the resource group already exists", func() {
			BeforeEach(func() {
				azureClient.CheckExistenceCall.Returns.Response = autorest.Response{
					Response: mocks.NewResponseWithStatus("some-message", 200),
				}
			})
			It("returns true", func() {
				exists, err := client.CheckExists("exact-same")
				Expect(err).NotTo(HaveOccurred())

				Expect(exists).To(BeTrue())
			})
		})

		Context("when the azure client returns an error", func() {
			BeforeEach(func() {
				azureClient.CheckExistenceCall.Returns.Error = errors.New("grape")
			})
			It("returns the error", func() {
				_, err := client.CheckExists("exact-same")
				Expect(err).To(MatchError("Check existence for resource group exact-same-bosh: grape"))
			})
		})
	})

	Describe("ValidateSafeToDelete", func() {
		var (
			azureClient *fakes.AzureVMsClient
			client      azure.Client
		)

		BeforeEach(func() {
			azureClient = &fakes.AzureVMsClient{}
			client = azure.NewClientWithInjectedVMsClient(azureClient)
		})

		Context("when the bosh director and jumpbox are the only vms in the network", func() {
			BeforeEach(func() {
				boshString := "bosh"
				jumpboxString := "jumpbox"

				azureClient.ListCall.Returns.Result = compute.VirtualMachineListResult{
					Value: &[]compute.VirtualMachine{
						{
							Tags: &map[string]*string{
								"job": &boshString,
							},
						},
						{
							Tags: &map[string]*string{
								"job": &jumpboxString,
							},
						},
					},
				}
			})

			It("does not return an error ", func() {
				err := client.ValidateSafeToDelete("", "some-env-id")
				Expect(err).NotTo(HaveOccurred())

				Expect(azureClient.ListCall.Receives.ResourceGroup).To(Equal("some-env-id-bosh"))
			})
		})

		Context("when some other bosh deployed vm exists in the network", func() {
			BeforeEach(func() {
				boshString := "bosh"
				jobString := "some-job"
				deploymentString := "some-deployment"
				vmNameString := "some-other-vm"

				azureClient.ListCall.Returns.Result = compute.VirtualMachineListResult{
					Value: &[]compute.VirtualMachine{
						{
							Tags: &map[string]*string{
								"job": &boshString,
							},
						},
						{
							Name: &vmNameString,
							Tags: &map[string]*string{
								"job":        &jobString,
								"deployment": &deploymentString,
							},
						},
					},
				}
			})

			It("returns a helpful error message", func() {
				err := client.ValidateSafeToDelete("", "some-env-id")
				Expect(err).To(MatchError(`bbl environment is not safe to delete; vms still exist in resource group: some-env-id-bosh (deployment: some-deployment): some-other-vm`))
			})
		})

		Context("when some other non-bosh deployed vm exists in the network", func() {
			BeforeEach(func() {
				vmNameString := "some-other-vm"
				azureClient.ListCall.Returns.Result = compute.VirtualMachineListResult{
					Value: &[]compute.VirtualMachine{
						{
							Name: &vmNameString,
							Tags: &map[string]*string{},
						},
					},
				}
			})

			It("returns a helpful error message", func() {
				err := client.ValidateSafeToDelete("", "some-env-id")
				Expect(err).To(MatchError(`bbl environment is not safe to delete; vms still exist in resource group: some-env-id-bosh: some-other-vm`))
			})
		})

		Context("failure cases", func() {
			Context("when azure client list instances fails", func() {
				BeforeEach(func() {
					azureClient.ListCall.Returns.Error = errors.New("passionfruit")
				})

				It("returns an error", func() {
					err := client.ValidateSafeToDelete("some-network", "some-env-id")
					Expect(err).To(MatchError("List instances: passionfruit"))
				})
			})
		})
	})
})
