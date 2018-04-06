package route53_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsroute53 "github.com/aws/aws-sdk-go/service/route53"
	"github.com/genevieve/leftovers/aws/route53"
	"github.com/genevieve/leftovers/aws/route53/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HostedZone", func() {
	var (
		hostedZone route53.HostedZone
		client     *fakes.HostedZonesClient
		id         *string
		name       *string
	)

	BeforeEach(func() {
		client = &fakes.HostedZonesClient{}
		id = aws.String("the-zone-id")
		name = aws.String("the-zone-name")

		hostedZone = route53.NewHostedZone(client, id, name)
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			client.ListResourceRecordSetsCall.Returns.Output = &awsroute53.ListResourceRecordSetsOutput{
				ResourceRecordSets: []*awsroute53.ResourceRecordSet{{
					Type: aws.String("something-else"),
				}},
			}
		})

		It("deletes the record sets and deletes the hosted zone", func() {
			err := hostedZone.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListResourceRecordSetsCall.CallCount).To(Equal(1))
			Expect(client.ListResourceRecordSetsCall.Receives.Input.HostedZoneId).To(Equal(id))

			Expect(client.ChangeResourceRecordSetsCall.CallCount).To(Equal(1))
			Expect(client.ChangeResourceRecordSetsCall.Receives.Input.HostedZoneId).To(Equal(id))
			Expect(client.ChangeResourceRecordSetsCall.Receives.Input.ChangeBatch.Changes[0].Action).To(Equal(aws.String("DELETE")))
			Expect(client.ChangeResourceRecordSetsCall.Receives.Input.ChangeBatch.Changes[0].ResourceRecordSet.Type).To(Equal(aws.String("something-else")))

			Expect(client.DeleteHostedZoneCall.CallCount).To(Equal(1))
			Expect(client.DeleteHostedZoneCall.Receives.Input.Id).To(Equal(id))
		})

		Context("when the resource record set is of type NS", func() {
			BeforeEach(func() {
				client.ListResourceRecordSetsCall.Returns.Output = &awsroute53.ListResourceRecordSetsOutput{
					ResourceRecordSets: []*awsroute53.ResourceRecordSet{{
						Type: aws.String("NS"),
					}},
				}
			})

			It("does not try to delete it", func() {
				err := hostedZone.Delete()
				Expect(err).NotTo(HaveOccurred())

				Expect(client.ListResourceRecordSetsCall.CallCount).To(Equal(1))
				Expect(client.ListResourceRecordSetsCall.Receives.Input.HostedZoneId).To(Equal(id))

				Expect(client.ChangeResourceRecordSetsCall.CallCount).To(Equal(0))

				Expect(client.DeleteHostedZoneCall.CallCount).To(Equal(1))
				Expect(client.DeleteHostedZoneCall.Receives.Input.Id).To(Equal(id))
			})
		})

		Context("when the resource record set is of type SOA", func() {
			BeforeEach(func() {
				client.ListResourceRecordSetsCall.Returns.Output = &awsroute53.ListResourceRecordSetsOutput{
					ResourceRecordSets: []*awsroute53.ResourceRecordSet{{
						Type: aws.String("SOA"),
					}},
				}
			})

			It("does not try to delete it", func() {
				err := hostedZone.Delete()
				Expect(err).NotTo(HaveOccurred())

				Expect(client.ListResourceRecordSetsCall.CallCount).To(Equal(1))
				Expect(client.ListResourceRecordSetsCall.Receives.Input.HostedZoneId).To(Equal(id))

				Expect(client.ChangeResourceRecordSetsCall.CallCount).To(Equal(0))

				Expect(client.DeleteHostedZoneCall.CallCount).To(Equal(1))
				Expect(client.DeleteHostedZoneCall.Receives.Input.Id).To(Equal(id))
			})
		})

		Context("when the client fails to list resource record sets", func() {
			BeforeEach(func() {
				client.ListResourceRecordSetsCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := hostedZone.Delete()
				Expect(err).To(MatchError("List Resource Record Sets: banana"))
			})
		})

		Context("when the client fails to delete resource record sets", func() {
			BeforeEach(func() {
				client.ChangeResourceRecordSetsCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := hostedZone.Delete()
				Expect(err).To(MatchError("Delete Resource Record Sets: banana"))
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
