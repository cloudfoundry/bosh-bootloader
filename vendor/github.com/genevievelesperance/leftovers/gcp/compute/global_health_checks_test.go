package compute_test

import (
	"errors"

	"github.com/genevievelesperance/leftovers/gcp/compute"
	"github.com/genevievelesperance/leftovers/gcp/compute/fakes"
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
			logger.PromptCall.Returns.Proceed = true
			client.ListGlobalHealthChecksCall.Returns.Output = &gcpcompute.HealthCheckList{
				Items: []*gcpcompute.HealthCheck{{
					Name: "banana-check",
				}},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for global health checks to delete", func() {
			list, err := globalHealthChecks.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListGlobalHealthChecksCall.CallCount).To(Equal(1))

			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete global health check banana-check?"))

			Expect(list).To(HaveLen(1))
			Expect(list).To(HaveKeyWithValue("banana-check", ""))
		})

		Context("when the client fails to list global health checks", func() {
			BeforeEach(func() {
				client.ListGlobalHealthChecksCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := globalHealthChecks.List(filter)
				Expect(err).To(MatchError("Listing global health checks: some error"))
			})
		})

		Context("when the health check name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := globalHealthChecks.List("grape")
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
				list, err := globalHealthChecks.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.CallCount).To(Equal(1))
				Expect(list).To(HaveLen(0))
			})
		})
	})

	Describe("Delete", func() {
		var list map[string]string

		BeforeEach(func() {
			list = map[string]string{"banana-check": ""}
		})

		It("deletes global health checks", func() {
			globalHealthChecks.Delete(list)

			Expect(client.DeleteGlobalHealthCheckCall.CallCount).To(Equal(1))
			Expect(client.DeleteGlobalHealthCheckCall.Receives.GlobalHealthCheck).To(Equal("banana-check"))

			Expect(logger.PrintfCall.Messages).To(Equal([]string{"SUCCESS deleting global health check banana-check\n"}))
		})

		Context("when the client fails to delete a global health check", func() {
			BeforeEach(func() {
				client.DeleteGlobalHealthCheckCall.Returns.Error = errors.New("some error")
			})

			It("logs the error", func() {
				globalHealthChecks.Delete(list)

				Expect(logger.PrintfCall.Messages).To(Equal([]string{"ERROR deleting global health check banana-check: some error\n"}))
			})
		})
	})
})
