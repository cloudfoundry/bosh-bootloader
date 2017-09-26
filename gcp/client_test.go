package gcp_test

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/gcp"
	compute "google.golang.org/api/compute/v1"

	. "github.com/onsi/ginkgo"
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
			var networkName string

			BeforeEach(func() {
				networkName = "some-network"
				directorName := "some-bosh-director"

				boshString := "bosh"
				boshInitString := "bosh-init"

				computeClient.ListInstancesCall.Returns.InstanceList = &compute.InstanceList{
					Items: []*compute.Instance{
						{
							Name: directorName,
							NetworkInterfaces: []*compute.NetworkInterface{
								{
									Network: "some-other-network",
								},
								{
									Network: fmt.Sprintf("http://some-host/%s", networkName),
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
									Network: "some-other-network",
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
				err := client.ValidateSafeToDelete(networkName, "some-env-id")

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when instances other than bosh director exist in network", func() {
			var (
				networkName    string
				vmName         string
				deploymentName string
				nonBOSHVMName  string
			)

			BeforeEach(func() {
				networkName = "some-network"
				directorName := "some-bosh-director"
				deploymentName = "some-deployment"
				vmName = "some-vm"
				nonBOSHVMName = "some-non-bosh-vm"

				boshString := "bosh"
				boshInitString := "bosh-init"

				computeClient.ListInstancesCall.Returns.InstanceList = &compute.InstanceList{
					Items: []*compute.Instance{
						{
							Name: directorName,
							NetworkInterfaces: []*compute.NetworkInterface{
								{
									Network: "some-other-network",
								},
								{
									Network: fmt.Sprintf("http://some-host/%s", networkName),
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
							Name: vmName,
							NetworkInterfaces: []*compute.NetworkInterface{
								{
									Network: "some-other-network",
								},
								{
									Network: fmt.Sprintf("http://some-host/%s", networkName),
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
							Name: nonBOSHVMName,
							NetworkInterfaces: []*compute.NetworkInterface{
								{
									Network: "some-other-network",
								},
								{
									Network: fmt.Sprintf("http://some-host/%s", networkName),
								},
							},
							Metadata: &compute.Metadata{
								Items: []*compute.MetadataItems{},
							},
						},
						{
							Name: "other-network-vm",
							NetworkInterfaces: []*compute.NetworkInterface{
								{
									Network: "some-other-network",
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
				err := client.ValidateSafeToDelete(networkName, "some-env-id")

				Expect(err).To(MatchError(fmt.Sprintf(`bbl environment is not safe to delete; vms still exist in network:
%s (deployment: %s)
%s (not managed by bosh)`, vmName, deploymentName, nonBOSHVMName)))
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
