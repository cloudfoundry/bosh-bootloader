package dns_test

import (
	"errors"

	gcpdns "google.golang.org/api/dns/v1"

	"github.com/genevieve/leftovers/gcp/dns"
	"github.com/genevieve/leftovers/gcp/dns/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ManagedZones", func() {
	var (
		client     *fakes.ManagedZonesClient
		recordSets *fakes.RecordSets
		logger     *fakes.Logger

		managedZones dns.ManagedZones
	)

	BeforeEach(func() {
		client = &fakes.ManagedZonesClient{}
		recordSets = &fakes.RecordSets{}
		logger = &fakes.Logger{}

		logger.PromptWithDetailsCall.Returns.Proceed = true

		managedZones = dns.NewManagedZones(client, recordSets, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			client.ListManagedZonesCall.Returns.Output = &gcpdns.ManagedZonesListResponse{
				ManagedZones: []*gcpdns.ManagedZone{{
					Name: "banana-managed-zone",
				}},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for managed zones to delete", func() {
			list, err := managedZones.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListManagedZonesCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("DNS Managed Zone"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-managed-zone"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list managed zones", func() {
			BeforeEach(func() {
				client.ListManagedZonesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := managedZones.List(filter)
				Expect(err).To(MatchError("Listing DNS Managed Zones: some error"))
			})
		})

		Context("when the managed zone name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := managedZones.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the user says no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not add it to the list", func() {
				list, err := managedZones.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
