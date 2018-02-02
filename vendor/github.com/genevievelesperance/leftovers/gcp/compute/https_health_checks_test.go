package compute_test

import (
	"errors"

	"github.com/genevievelesperance/leftovers/gcp/compute"
	"github.com/genevievelesperance/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("HttpsHealthChecks", func() {
	var (
		client *fakes.HttpsHealthChecksClient
		logger *fakes.Logger

		httpsHealthChecks compute.HttpsHealthChecks
	)

	BeforeEach(func() {
		client = &fakes.HttpsHealthChecksClient{}
		logger = &fakes.Logger{}

		httpsHealthChecks = compute.NewHttpsHealthChecks(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptCall.Returns.Proceed = true
			client.ListHttpsHealthChecksCall.Returns.Output = &gcpcompute.HttpsHealthCheckList{
				Items: []*gcpcompute.HttpsHealthCheck{{
					Name: "banana-check",
				}},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for https health checks to delete", func() {
			list, err := httpsHealthChecks.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListHttpsHealthChecksCall.CallCount).To(Equal(1))

			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete https health check banana-check?"))

			Expect(list).To(HaveLen(1))
			Expect(list).To(HaveKeyWithValue("banana-check", ""))
		})

		Context("when the client fails to list https health checks", func() {
			BeforeEach(func() {
				client.ListHttpsHealthChecksCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := httpsHealthChecks.List(filter)
				Expect(err).To(MatchError("Listing https health checks: some error"))
			})
		})

		Context("when the health check name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := httpsHealthChecks.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.CallCount).To(Equal(0))
				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the user says no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptCall.Returns.Proceed = false
			})

			It("does not add it to the list ", func() {
				list, err := httpsHealthChecks.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})

	Describe("Delete", func() {
		var list map[string]string

		BeforeEach(func() {
			list = map[string]string{"banana-check": ""}
		})

		It("deletes https health checks", func() {
			httpsHealthChecks.Delete(list)

			Expect(client.DeleteHttpsHealthCheckCall.CallCount).To(Equal(1))
			Expect(client.DeleteHttpsHealthCheckCall.Receives.HttpsHealthCheck).To(Equal("banana-check"))

			Expect(logger.PrintfCall.Messages).To(Equal([]string{"SUCCESS deleting https health check banana-check\n"}))
		})

		Context("when the client fails to delete the https health check", func() {
			BeforeEach(func() {
				client.DeleteHttpsHealthCheckCall.Returns.Error = errors.New("some error")
			})

			It("logs the error", func() {
				httpsHealthChecks.Delete(list)

				Expect(logger.PrintfCall.Messages).To(Equal([]string{"ERROR deleting https health check banana-check: some error\n"}))
			})
		})
	})
})
