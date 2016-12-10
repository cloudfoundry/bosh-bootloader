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

var _ = Describe("NetworkInstancesRetriever", func() {
	var (
		client                    *fakes.GCPClient
		gcpClientProvider         *fakes.GCPClientProvider
		logger                    *fakes.Logger
		networkInstancesRetriever gcp.NetworkInstancesRetriever
	)
	BeforeEach(func() {
		gcpClientProvider = &fakes.GCPClientProvider{}
		client = &fakes.GCPClient{}
		logger = &fakes.Logger{}
		gcpClientProvider.ClientCall.Returns.Client = client
		networkInstancesRetriever = gcp.NewNetworkInstancesRetriever(gcpClientProvider, logger)
	})

	It("returns a list of instances within a gcp network", func() {
		network := "some-network"
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
				},
				{
					Name: "some-other-vm",
					NetworkInterfaces: []*compute.NetworkInterface{
						{
							Network: "http://some-host/some-other-network",
						},
					},
				},
			},
		}
		instances, err := networkInstancesRetriever.List("some-project-id", "some-zone", network)
		Expect(err).NotTo(HaveOccurred())

		Expect(gcpClientProvider.ClientCall.CallCount).To(Equal(1))

		Expect(client.ListInstancesCall.Receives.ProjectID).To(Equal("some-project-id"))
		Expect(client.ListInstancesCall.Receives.Zone).To(Equal("some-zone"))
		Expect(instances).To(Equal([]string{"some-vm"}))
	})

	Context("failure cases", func() {
		It("returns an error when gcp client list instances fails", func() {
			client.ListInstancesCall.Returns.Error = errors.New("fails to list instances")
			_, err := networkInstancesRetriever.List("some-project-id", "some-zone", "some-network")
			Expect(err).To(MatchError("fails to list instances"))
		})
	})
})
