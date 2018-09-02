package route53_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevieve/leftovers/aws/route53"
	"github.com/genevieve/leftovers/aws/route53/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HostedZone", func() {
	var (
		client     *fakes.HostedZonesClient
		recordSets *fakes.RecordSets
		id         *string
		name       *string

		hostedZone route53.HostedZone
	)

	BeforeEach(func() {
		client = &fakes.HostedZonesClient{}
		recordSets = &fakes.RecordSets{}
		id = aws.String("the-zone-id")
		name = aws.String("the-zone-name")

		hostedZone = route53.NewHostedZone(client, id, name, recordSets)
	})

	Describe("Delete", func() {
		It("deletes the record sets and deletes the hosted zone", func() {
			err := hostedZone.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(recordSets.GetCall.CallCount).To(Equal(1))
			Expect(recordSets.GetCall.Receives.HostedZoneId).To(Equal(id))

			Expect(recordSets.DeleteCall.CallCount).To(Equal(1))
			Expect(recordSets.DeleteCall.Receives.HostedZoneId).To(Equal(id))

			Expect(client.DeleteHostedZoneCall.CallCount).To(Equal(1))
			Expect(client.DeleteHostedZoneCall.Receives.Input.Id).To(Equal(id))
		})

		Context("when record sets fails to get", func() {
			BeforeEach(func() {
				recordSets.GetCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := hostedZone.Delete()
				Expect(err).To(MatchError("Get Record Sets: banana"))
			})
		})

		Context("when record sets fails to delete", func() {
			BeforeEach(func() {
				recordSets.DeleteCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := hostedZone.Delete()
				Expect(err).To(MatchError("Delete Record Sets: banana"))
			})
		})

		Context("when the client fails to delete the zone", func() {
			BeforeEach(func() {
				client.DeleteHostedZoneCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := hostedZone.Delete()
				Expect(err).To(MatchError("Delete: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(hostedZone.Name()).To(Equal("the-zone-name"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(hostedZone.Type()).To(Equal("Route53 Hosted Zone"))
		})
	})
})
