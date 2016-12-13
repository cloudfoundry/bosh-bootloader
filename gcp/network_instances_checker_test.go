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
		gcpClientProvider       *fakes.GCPClientProvider
		networkInstancesChecker gcp.NetworkInstancesChecker
	)

	Describe("ValidateSafeToDelete", func() {
		BeforeEach(func() {
			gcpClientProvider = &fakes.GCPClientProvider{}
			client = &fakes.GCPClient{}
			gcpClientProvider.ClientCall.Returns.Client = client
			networkInstancesChecker = gcp.NewNetworkInstancesChecker(gcpClientProvider)
		})

		It("returns helpful error message when instances instances other than bosh director exist in network", func() {
			network := "some-network"
			boshInit := "bosh-init"
			directorName := "some-bosh-director"
			deployment := "some-deployment"

			client.ListInstancesCall.Returns.InstanceList = &compute.InstanceList{
				Items: []*compute.Instance{
					{
						Name: "some-vm",
						NetworkInterfaces: []*compute.NetworkInterface{
							{
								Network: "some-other-network",
							},
							{
								Network: fmt.Sprintf("http://some-host/%s", network),
							},
						},
						Metadata: &compute.Metadata{
							Items: []*compute.MetadataItems{
								{
									Key:   "director",
									Value: &boshInit,
								},
							},
						},
					},
					{
						Name: "some-other-vm",
						NetworkInterfaces: []*compute.NetworkInterface{
							{
								Network: fmt.Sprintf("http://some-host/%s", network),
							},
						},
						Metadata: &compute.Metadata{
							Items: []*compute.MetadataItems{
								{
									Key:   "director",
									Value: &directorName,
								},
								{
									Key:   "deployment",
									Value: &deployment,
								},
							},
						},
					},
					{
						Name: "non-bosh-vm",
						NetworkInterfaces: []*compute.NetworkInterface{
							{
								Network: fmt.Sprintf("http://some-host/%s", network),
							},
						},
						Metadata: &compute.Metadata{
							Items: []*compute.MetadataItems{},
						},
					},
					{
						Name: "some-other-bosh-director-vm",
						NetworkInterfaces: []*compute.NetworkInterface{
							{
								Network: "another-network",
							},
						},
						Metadata: &compute.Metadata{
							Items: []*compute.MetadataItems{
								{
									Key:   "director",
									Value: &directorName,
								},
							},
						},
					},
				},
			}
			err := networkInstancesChecker.ValidateSafeToDelete(network)

			Expect(gcpClientProvider.ClientCall.CallCount).To(Equal(1))

			Expect(err).To(MatchError(`bbl environment is not safe to delete; vms still exist in network:
some-other-vm (deployment: some-deployment)
non-bosh-vm (not managed by bosh)`))
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
