package dns_test

import (
	"errors"

	gcpdns "google.golang.org/api/dns/v1"

	"github.com/genevieve/leftovers/gcp/dns"
	"github.com/genevieve/leftovers/gcp/dns/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RecordSets", func() {
	var (
		client *fakes.RecordSetsClient

		recordSets dns.RecordSets
	)

	BeforeEach(func() {
		client = &fakes.RecordSetsClient{}

		recordSets = dns.NewRecordSets(client)
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			client.ListRecordSetsCall.Returns.Output = &gcpdns.ResourceRecordSetsListResponse{
				Rrsets: []*gcpdns.ResourceRecordSet{{
					Name: "banana-record",
					Type: "not-ns-or-soa",
				}},
			}
		})

		It("lists and deletes record sets for the managed zone", func() {
			err := recordSets.Delete("the-zone")
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListRecordSetsCall.CallCount).To(Equal(1))
			Expect(client.ListRecordSetsCall.Receives.ManagedZone).To(Equal("the-zone"))

			Expect(client.DeleteRecordSetsCall.CallCount).To(Equal(1))
			Expect(client.DeleteRecordSetsCall.Receives.ManagedZone).To(Equal("the-zone"))
			Expect(client.DeleteRecordSetsCall.Receives.Change.Deletions).To(HaveLen(1))
			Expect(client.DeleteRecordSetsCall.Receives.Change.Deletions[0].Name).To(Equal("banana-record"))
			Expect(client.DeleteRecordSetsCall.Receives.Change.Deletions[0].Type).To(Equal("not-ns-or-soa"))
		})

		Context("when the record type is NS", func() {
			BeforeEach(func() {
				client.ListRecordSetsCall.Returns.Output = &gcpdns.ResourceRecordSetsListResponse{
					Rrsets: []*gcpdns.ResourceRecordSet{{
						Name: "banana-record",
						Type: "NS",
					}},
				}
			})

			//Zone must contain exactly one resource record set of type NS at the apex.
			It("does not delete it", func() {
				err := recordSets.Delete("the-zone")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DeleteRecordSetsCall.CallCount).To(Equal(0))
			})
		})

		Context("when the record type is SOA", func() {
			BeforeEach(func() {
				client.ListRecordSetsCall.Returns.Output = &gcpdns.ResourceRecordSetsListResponse{
					Rrsets: []*gcpdns.ResourceRecordSet{{
						Name: "banana-record",
						Type: "SOA",
					}},
				}
			})

			//Zone must contain exactly one resource record set of type SOA at the apex.
			It("does not delete it", func() {
				err := recordSets.Delete("the-zone")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DeleteRecordSetsCall.CallCount).To(Equal(0))
			})
		})

		Context("when the client fails to list record sets for the zone", func() {
			BeforeEach(func() {
				client.ListRecordSetsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				err := recordSets.Delete("the-zone")
				Expect(err).To(MatchError("Listing DNS Record Sets: some error"))
			})
		})

		Context("when the client fails to delete the record sets", func() {
			BeforeEach(func() {
				client.DeleteRecordSetsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				err := recordSets.Delete("the-zone")
				Expect(err).To(MatchError("Delete record sets: some error"))
			})
		})
	})
})
