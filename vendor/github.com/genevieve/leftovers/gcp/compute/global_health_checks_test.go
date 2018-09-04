package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("GlobalHealthChecks", func() {
	var (
		client *fakes.GlobalHealthChecksClient
		logger *fakes.Logger

		globalHealthChecks compute.GlobalHealthChecks
	)

	BeforeEach(func() {
		client = &fakes.GlobalHealthChecksClient{}
		logger = &fakes.Logger{}

		globalHealthChecks = compute.NewGlobalHealthChecks(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListGlobalHealthChecksCall.Returns.Output = []*gcpcompute.HealthCheck{{
				Name: "banana-check",
			}}
			filter = "banana"
		})

		It("lists, filters, and prompts for global health checks to delete", func() {
			list, err := globalHealthChecks.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListGlobalHealthChecksCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Global Health Check"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-check"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list global health checks", func() {
			BeforeEach(func() {
				client.ListGlobalHealthChecksCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := globalHealthChecks.List(filter)
				Expect(err).To(MatchError("List Global Health Checks: some error"))
			})
		})

		Context("when the health check name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := globalHealthChecks.List("grape")
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
				list, err := globalHealthChecks.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(list).To(HaveLen(0))
			})
		})
	})
})
