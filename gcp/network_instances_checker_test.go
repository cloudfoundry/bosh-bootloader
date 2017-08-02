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

var _ = Describe("network instances checker", func() {
	var (
		client                  *fakes.GCPClient
		networkInstancesChecker gcp.NetworkInstancesChecker
	)

	Describe("ValidateSafeToDelete", func() {
		BeforeEach(func() {
			client = &fakes.GCPClient{}
			networkInstancesChecker = gcp.NewNetworkInstancesChecker(client)
		})

		It("does not return an error when the bosh director is the only vm on the network", func() {
			networkName := "some-network"
			directorName := "some-bosh-director"

			boshString := "bosh"
			boshInitString := "bosh-init"

			client.ListInstancesCall.Returns.InstanceList = &compute.InstanceList{
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

			err := networkInstancesChecker.ValidateSafeToDelete(networkName)

			Expect(err).NotTo(HaveOccurred())
		})

		It("returns helpful error message when instances other than bosh director exist in network", func() {
			networkName := "some-network"
			directorName := "some-bosh-director"
			deploymentName := "some-deployment"
			vmName := "some-vm"
			nonBOSHVMName := "some-non-bosh-vm"

			boshString := "bosh"
			boshInitString := "bosh-init"

			client.ListInstancesCall.Returns.InstanceList = &compute.InstanceList{
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

			err := networkInstancesChecker.ValidateSafeToDelete(networkName)

			Expect(err).To(MatchError(fmt.Sprintf(`bbl environment is not safe to delete; vms still exist in network:
%s (deployment: %s)
%s (not managed by bosh)`, vmName, deploymentName, nonBOSHVMName)))
		})

		Context("failure cases", func() {
			It("returns an error when gcp client list instances fails", func() {
				client.ListInstancesCall.Returns.Error = errors.New("fails to list instances")
				err := networkInstancesChecker.ValidateSafeToDelete("some-network")
				Expect(err).To(MatchError("fails to list instances"))
			})
		})
	})

})
