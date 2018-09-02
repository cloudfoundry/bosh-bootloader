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

var _ = Describe("HostedZones", func() {
	var (
		client     *fakes.HostedZonesClient
		logger     *fakes.Logger
		recordSets *fakes.RecordSets

		hostedZones route53.HostedZones
	)

	BeforeEach(func() {
		client = &fakes.HostedZonesClient{}
		logger = &fakes.Logger{}

		hostedZones = route53.NewHostedZones(client, logger, recordSets)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListHostedZonesCall.Returns.Output = &awsroute53.ListHostedZonesOutput{
				HostedZones: []*awsroute53.HostedZone{{
					Id:   aws.String("the-id"),
					Name: aws.String("banana"),
				}},
			}
			filter = "ban"
		})

		It("returns a list of route53 hosted zones to delete", func() {
			items, err := hostedZones.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListHostedZonesCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Route53 Hosted Zone"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the client fails to list hosted zones", func() {
			BeforeEach(func() {
				client.ListHostedZonesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := hostedZones.List(filter)
				Expect(err).To(MatchError("List Route53 Hosted Zones: some error"))
			})
		})

		Context("when the hosted zone name does not contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := hostedZones.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not return it in the list", func() {
				items, err := hostedZones.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
