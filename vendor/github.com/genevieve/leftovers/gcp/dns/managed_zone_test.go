package dns_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/dns"
	"github.com/genevieve/leftovers/gcp/dns/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ManagedZone", func() {
	var (
		client     *fakes.ManagedZonesClient
		recordSets *fakes.RecordSets
		name       string

		managedZone dns.ManagedZone
	)

	BeforeEach(func() {
		client = &fakes.ManagedZonesClient{}
		recordSets = &fakes.RecordSets{}
		name = "banana"

		managedZone = dns.NewManagedZone(client, recordSets, name)
	})

	Describe("Delete", func() {
		It("deletes the managed zone", func() {
			err := managedZone.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(recordSets.DeleteCall.CallCount).To(Equal(1))
			Expect(recordSets.DeleteCall.Receives.ManagedZone).To(Equal(name))

			Expect(client.DeleteManagedZoneCall.CallCount).To(Equal(1))
			Expect(client.DeleteManagedZoneCall.Receives.ManagedZone).To(Equal(name))
		})

		Context("when the client fails to delete the record sets", func() {
			BeforeEach(func() {
				recordSets.DeleteCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := managedZone.Delete()
				Expect(err).To(MatchError("Delete record sets: the-error"))
			})
		})

		Context("when the client fails to delete the managed zone", func() {
			BeforeEach(func() {
				client.DeleteManagedZoneCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := managedZone.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(managedZone.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(managedZone.Type()).To(Equal("DNS Managed Zone"))
		})
	})
})
