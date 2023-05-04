package gcp_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/gcp"
	compute "google.golang.org/api/compute/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var (
		computeClient *fakes.GCPComputeClient
		client        gcp.Client
	)

	Describe("ValidateSafeToDelete", func() {
		BeforeEach(func() {
			computeClient = &fakes.GCPComputeClient{}
			client = gcp.NewClientWithInjectedComputeClient(computeClient, "some-project-id", "some-zone")
		})

		Context("when the bosh director is the only vm on the network", func() {

			BeforeEach(func() {
				boshString := "bosh"
				boshInitString := "bosh-init"

				computeClient.ListInstancesCall.Returns.InstanceList = &compute.InstanceList{
					Items: []*compute.Instance{
						{
							Name: "bosh-director",
							NetworkInterfaces: []*compute.NetworkInterface{
								{
									Network: "network-name",
								},
							},
							Metadata: &compute.Metadata{
								Items: []*compute.MetadataItems{
									{
										Key:   "deployment",
										Value: &boshString,
									},
									{
										Key:   "director",
										Value: &boshInitString,
									},
								},
							},
						},
						{
							Name: "other-network-vm",
							NetworkInterfaces: []*compute.NetworkInterface{
								{
									Network: "other-network-name",
								},
							},
							Metadata: &compute.Metadata{
								Items: []*compute.MetadataItems{},
							},
						},
					},
				}
			})

			It("does not return an error ", func() {
				err := client.ValidateSafeToDelete("network-name", "some-env-id")

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when vms other than the bosh director exist in network", func() {
			BeforeEach(func() {
				directorName := "some-bosh-director"
				deploymentName := "some-deployment"

				computeClient.ListInstancesCall.Returns.InstanceList = &compute.InstanceList{
					Items: []*compute.Instance{
						{
							Name: "bosh-managed-vm",
							NetworkInterfaces: []*compute.NetworkInterface{
								{
									Network: "network-name",
								},
							},
							Metadata: &compute.Metadata{
								Items: []*compute.MetadataItems{
									{
										Key:   "deployment",
										Value: &deploymentName,
									},
									{
										Key:   "director",
										Value: &directorName,
									},
								},
							},
						},
						{
							Name: "not-a-bosh-managed-vm",
							NetworkInterfaces: []*compute.NetworkInterface{
								{
									Network: "network-name",
								},
							},
							Metadata: &compute.Metadata{
								Items: []*compute.MetadataItems{},
							},
						},
					},
				}
			})

			It("returns a helpful error message", func() {
				err := client.ValidateSafeToDelete("network-name", "some-env-id")

				Expect(err).To(MatchError(`bbl environment is not safe to delete; vms still exist in network:
bosh-managed-vm (deployment: some-deployment)
not-a-bosh-managed-vm (not managed by bosh)`))
			})
		})

		Context("failure cases", func() {
			Context("when gcp client list instances fails", func() {
				BeforeEach(func() {
					computeClient.ListInstancesCall.Returns.Error = errors.New("fails to list instances")
				})

				It("returns an error", func() {
					err := client.ValidateSafeToDelete("some-network", "some-env-id")
					Expect(err).To(MatchError("fails to list instances"))
				})
			})
		})
	})
})
