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

var _ = Describe("RecordSets", func() {
	var (
		client         *fakes.RecordSetsClient
		hostedZoneId   *string
		hostedZoneName string

		recordSets route53.RecordSets
	)

	BeforeEach(func() {
		client = &fakes.RecordSetsClient{}
		hostedZoneId = aws.String("zone-id")
		hostedZoneName = "zone-name"

		recordSets = route53.NewRecordSets(client)
	})

	Describe("Get", func() {
		BeforeEach(func() {
			client.ListResourceRecordSetsCall.Returns = []fakes.ListResourceRecordSetsCallReturn{{
				Output: &awsroute53.ListResourceRecordSetsOutput{
					ResourceRecordSets: []*awsroute53.ResourceRecordSet{{
						Name: aws.String("the-name"),
						Type: aws.String("something-else"),
					}},
					IsTruncated: aws.Bool(false),
				}},
			}
		})

		It("gets the record sets", func() {
			records, err := recordSets.Get(hostedZoneId)
			Expect(err).NotTo(HaveOccurred())

			Expect(records).To(HaveLen(1))

			Expect(client.ListResourceRecordSetsCall.CallCount).To(Equal(1))
			Expect(client.ListResourceRecordSetsCall.Receives[0].Input.HostedZoneId).To(Equal(hostedZoneId))
		})

		Context("when there are pages of record sets", func() {
			BeforeEach(func() {
				client.ListResourceRecordSetsCall.Returns = []fakes.ListResourceRecordSetsCallReturn{
					{
						Output: &awsroute53.ListResourceRecordSetsOutput{
							ResourceRecordSets: []*awsroute53.ResourceRecordSet{{
								Type: aws.String("something-else"),
							}},
							NextRecordName: aws.String("one-more-thing"),
							IsTruncated:    aws.Bool(true),
						},
					},
					{
						Output: &awsroute53.ListResourceRecordSetsOutput{
							ResourceRecordSets: []*awsroute53.ResourceRecordSet{{
								Type: aws.String("one-more-thing"),
							}},
							IsTruncated: aws.Bool(false),
						},
					},
				}
			})

			It("loops over the list request", func() {
				records, err := recordSets.Get(hostedZoneId)
				Expect(err).NotTo(HaveOccurred())

				Expect(records).To(HaveLen(2))

				Expect(client.ListResourceRecordSetsCall.CallCount).To(Equal(2))
				Expect(client.ListResourceRecordSetsCall.Receives[0].Input.HostedZoneId).To(Equal(hostedZoneId))
				Expect(client.ListResourceRecordSetsCall.Receives[0].Input.StartRecordName).To(BeNil())
				Expect(client.ListResourceRecordSetsCall.Receives[1].Input.StartRecordName).To(Equal(aws.String("one-more-thing")))
			})
		})

		Context("when the client fails to list resource record sets", func() {
			BeforeEach(func() {
				client.ListResourceRecordSetsCall.Returns = []fakes.ListResourceRecordSetsCallReturn{{Error: errors.New("banana")}}
			})

			It("returns the error", func() {
				_, err := recordSets.Get(hostedZoneId)
				Expect(err).To(MatchError("List Resource Record Sets: banana"))
			})
		})
	})

	Describe("Delete", func() {
		var records []*awsroute53.ResourceRecordSet

		BeforeEach(func() {
			records = []*awsroute53.ResourceRecordSet{{
				Name: aws.String(hostedZoneName),
				Type: aws.String("something-else"),
			}}
		})

		It("deletes the record sets", func() {
			err := recordSets.Delete(hostedZoneId, hostedZoneName, records)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ChangeResourceRecordSetsCall.CallCount).To(Equal(1))
			Expect(client.ChangeResourceRecordSetsCall.Receives.Input.HostedZoneId).To(Equal(hostedZoneId))
			Expect(client.ChangeResourceRecordSetsCall.Receives.Input.ChangeBatch.Changes[0].Action).To(Equal(aws.String("DELETE")))
			Expect(client.ChangeResourceRecordSetsCall.Receives.Input.ChangeBatch.Changes[0].ResourceRecordSet.Type).To(Equal(aws.String("something-else")))
		})

		Context("when the resource record set is of type NS", func() {
			BeforeEach(func() {
				records = []*awsroute53.ResourceRecordSet{{
					Name: aws.String(hostedZoneName),
					Type: aws.String("NS"),
				}}
			})

			It("does not try to delete it", func() {
				err := recordSets.Delete(hostedZoneId, hostedZoneName, records)
				Expect(err).NotTo(HaveOccurred())

				Expect(client.ChangeResourceRecordSetsCall.CallCount).To(Equal(0))
			})
		})

		Context("when the resource record set is of type SOA", func() {
			BeforeEach(func() {
				records = []*awsroute53.ResourceRecordSet{{
					Name: aws.String(hostedZoneName),
					Type: aws.String("SOA"),
				}}
			})

			It("does not try to delete it", func() {
				err := recordSets.Delete(hostedZoneId, hostedZoneName, records)
				Expect(err).NotTo(HaveOccurred())

				Expect(client.ChangeResourceRecordSetsCall.CallCount).To(Equal(0))
			})
		})

		Context("when the client fails to delete resource record sets", func() {
			BeforeEach(func() {
				client.ChangeResourceRecordSetsCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := recordSets.Delete(hostedZoneId, hostedZoneName, records)
				Expect(err).To(MatchError("Delete Resource Record Sets: banana"))
			})
		})
	})
})
