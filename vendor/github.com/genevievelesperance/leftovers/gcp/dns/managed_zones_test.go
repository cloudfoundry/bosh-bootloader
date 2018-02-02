package dns_test

import (
	"errors"

	gcpdns "google.golang.org/api/dns/v1"

	"github.com/genevievelesperance/leftovers/gcp/dns"
	"github.com/genevievelesperance/leftovers/gcp/dns/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ManagedZones", func() {
	var (
		client *fakes.ManagedZonesClient
		logger *fakes.Logger

		managedZones dns.ManagedZones
	)

	BeforeEach(func() {
		client = &fakes.ManagedZonesClient{}
		logger = &fakes.Logger{}

		logger.PromptCall.Returns.Proceed = true

		managedZones = dns.NewManagedZones(client, logger)
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

			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete managed zone banana-managed-zone?"))

			Expect(list).To(HaveLen(1))
			Expect(list).To(HaveKeyWithValue("banana-managed-zone", ""))
		})

		Context("when the client fails to list managed zones", func() {
			BeforeEach(func() {
				client.ListManagedZonesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := managedZones.List(filter)
				Expect(err).To(MatchError("Listing managed zones: some error"))
			})
		})

		Context("when the managed zone name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := managedZones.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.CallCount).To(Equal(0))
				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the user says no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptCall.Returns.Proceed = false
			})

			It("does not add it to the list", func() {
				list, err := managedZones.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})

	Describe("Delete", func() {
		var list map[string]string

		BeforeEach(func() {
			list = map[string]string{"banana-managed-zone": ""}
		})

		It("deletes managed zones", func() {
			managedZones.Delete(list)

			Expect(client.DeleteManagedZoneCall.CallCount).To(Equal(1))
			Expect(client.DeleteManagedZoneCall.Receives.ManagedZone).To(Equal("banana-managed-zone"))

			Expect(logger.PrintfCall.Messages).To(Equal([]string{"SUCCESS deleting managed zone banana-managed-zone\n"}))
		})

		Context("when the client fails to delete the managed zone", func() {
			BeforeEach(func() {
				client.DeleteManagedZoneCall.Returns.Error = errors.New("some error")
			})

			It("logs the error", func() {
				managedZones.Delete(list)

				Expect(logger.PrintfCall.Messages).To(Equal([]string{"ERROR deleting managed zone banana-managed-zone: some error\n"}))
			})
		})
	})
})
