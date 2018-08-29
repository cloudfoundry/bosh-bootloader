package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("BackendServices", func() {
	var (
		client *fakes.BackendServicesClient
		logger *fakes.Logger

		backendServices compute.BackendServices
	)

	BeforeEach(func() {
		client = &fakes.BackendServicesClient{}
		logger = &fakes.Logger{}

		backendServices = compute.NewBackendServices(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			filter = "banana"
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListBackendServicesCall.Returns.Output = []*gcpcompute.BackendService{{
				Name: "banana-backend-service",
			}}
		})

		It("lists, filters, and prompts for backend services to delete", func() {
			list, err := backendServices.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListBackendServicesCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Backend Service"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-backend-service"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list backend services", func() {
			BeforeEach(func() {
				client.ListBackendServicesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := backendServices.List(filter)
				Expect(err).To(MatchError("List Backend Services: some error"))
			})
		})

		Context("when the backend service name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := backendServices.List("grape")
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
				list, err := backendServices.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
