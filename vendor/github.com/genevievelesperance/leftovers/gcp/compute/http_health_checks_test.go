package compute_test

import (
	"errors"

	"github.com/genevievelesperance/leftovers/gcp/compute"
	"github.com/genevievelesperance/leftovers/gcp/compute/fakes"
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
			logger.PromptCall.Returns.Proceed = true
			client.ListHttpHealthChecksCall.Returns.Output = &gcpcompute.HttpHealthCheckList{
				Items: []*gcpcompute.HttpHealthCheck{{
					Name: "banana-check",
				}},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for http health checks to delete", func() {
			list, err := httpHealthChecks.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListHttpHealthChecksCall.CallCount).To(Equal(1))

			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete http health check banana-check?"))

			Expect(list).To(HaveLen(1))
			Expect(list).To(HaveKeyWithValue("banana-check", ""))
		})

		Context("when the client fails to list http health checks", func() {
			BeforeEach(func() {
				client.ListHttpHealthChecksCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := httpHealthChecks.List(filter)
				Expect(err).To(MatchError("Listing http health checks: some error"))
			})
		})

		Context("when the health check name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := httpHealthChecks.List("grape")
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
				list, err := httpHealthChecks.List(filter)
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

		It("deletes http health checks", func() {
			httpHealthChecks.Delete(list)

			Expect(client.DeleteHttpHealthCheckCall.CallCount).To(Equal(1))
			Expect(client.DeleteHttpHealthCheckCall.Receives.HttpHealthCheck).To(Equal("banana-check"))

			Expect(logger.PrintfCall.Messages).To(Equal([]string{"SUCCESS deleting http health check banana-check\n"}))
		})

		Context("when the client fails to delete a http health check", func() {
			BeforeEach(func() {
				client.DeleteHttpHealthCheckCall.Returns.Error = errors.New("some error")
			})

			It("logs the error", func() {
				httpHealthChecks.Delete(list)

				Expect(logger.PrintfCall.Messages).To(Equal([]string{"ERROR deleting http health check banana-check: some error\n"}))
			})
		})
	})
})
