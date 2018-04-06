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

var _ = Describe("HealthChecks", func() {
	var (
		client *fakes.HealthChecksClient
		logger *fakes.Logger

		healthChecks route53.HealthChecks
	)

	BeforeEach(func() {
		client = &fakes.HealthChecksClient{}
		logger = &fakes.Logger{}

		healthChecks = route53.NewHealthChecks(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListHealthChecksCall.Returns.Output = &awsroute53.ListHealthChecksOutput{
				HealthChecks: []*awsroute53.HealthCheck{{
					Id: aws.String("the-id"),
				}},
			}
			filter = "the"
		})

		It("returns a list of route53 health checks to delete", func() {
			items, err := healthChecks.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListHealthChecksCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Route53 Health Check"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("the-id"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the client fails to list health checks", func() {
			BeforeEach(func() {
				client.ListHealthChecksCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := healthChecks.List(filter)
				Expect(err).To(MatchError("List Route53 Health Checks: some error"))
			})
		})

		Context("when the health check name does not contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := healthChecks.List("kiwi")
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
				items, err := healthChecks.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
