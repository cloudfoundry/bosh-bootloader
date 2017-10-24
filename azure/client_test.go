package azure_test

import (
	"errors"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/cloudfoundry/bosh-bootloader/azure"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var (
		// computeClient *fakes.GCPComputeClient
		azureClient *fakes.AzureVMsClient
		client      azure.Client
	)

	Describe("ValidateSafeToDelete", func() {
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
						compute.VirtualMachine{
							Tags: &map[string]*string{
								"job": &boshString,
							},
						},
						compute.VirtualMachine{
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
						compute.VirtualMachine{
							Tags: &map[string]*string{
								"job": &boshString,
							},
						},
						compute.VirtualMachine{
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
						compute.VirtualMachine{
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
