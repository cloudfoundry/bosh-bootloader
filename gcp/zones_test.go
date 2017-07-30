package gcp_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/gcp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("zones", func() {
	var (
		client            *fakes.GCPClient
		gcpClientProvider *fakes.GCPClientProvider
		zones             gcp.Zones
	)

	BeforeEach(func() {
		gcpClientProvider = &fakes.GCPClientProvider{}
		client = &fakes.GCPClient{}
		gcpClientProvider.ClientCall.Returns.Client = client

		client.GetZonesCall.Returns.Zones = []string{"zone-a", "zone-b"}

		zones = gcp.NewZones(gcpClientProvider)
	})

	Describe("get", func() {
		It("returns a list of zones for a given region", func() {
			actualZones, err := zones.Get("region-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(actualZones).To(Equal([]string{"zone-a", "zone-b"}))
		})

		Context("when gcp client get zones fails", func() {
			It("returns the error", func() {
				client.GetZonesCall.Returns.Error = errors.New("failed to get zones")
				_, err := zones.Get("some-region")
				Expect(err).To(MatchError("failed to get zones"))
			})
		})
	})
})
