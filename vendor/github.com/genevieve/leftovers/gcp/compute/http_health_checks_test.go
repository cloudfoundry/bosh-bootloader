package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("HttpHealthChecks", func() {
	var (
		client *fakes.HttpHealthChecksClient
		logger *fakes.Logger

		httpHealthChecks compute.HttpHealthChecks
	)

	BeforeEach(func() {
		client = &fakes.HttpHealthChecksClient{}
		logger = &fakes.Logger{}

		httpHealthChecks = compute.NewHttpHealthChecks(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListHttpHealthChecksCall.Returns.Output = []*gcpcompute.HttpHealthCheck{{
				Name: "banana-check",
			}}
			filter = "banana"
		})

		It("lists, filters, and prompts for http health checks to delete", func() {
			list, err := httpHealthChecks.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListHttpHealthChecksCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Http Health Check"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-check"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list http health checks", func() {
			BeforeEach(func() {
				client.ListHttpHealthChecksCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := httpHealthChecks.List(filter)
				Expect(err).To(MatchError("List Http Health Checks: some error"))
			})
		})

		Context("when the health check name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := httpHealthChecks.List("grape")
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
				list, err := httpHealthChecks.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(list).To(HaveLen(0))
			})
		})
	})
})
