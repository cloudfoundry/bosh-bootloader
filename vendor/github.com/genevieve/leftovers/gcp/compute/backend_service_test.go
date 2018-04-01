package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BackendService", func() {
	var (
		client *fakes.BackendServicesClient
		name   string

		backendService compute.BackendService
	)

	BeforeEach(func() {
		client = &fakes.BackendServicesClient{}
		name = "banana"

		backendService = compute.NewBackendService(client, name)
	})

	Describe("Delete", func() {
		It("deletes the backend service", func() {
			err := backendService.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteBackendServiceCall.CallCount).To(Equal(1))
			Expect(client.DeleteBackendServiceCall.Receives.BackendService).To(Equal(name))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteBackendServiceCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := backendService.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(backendService.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(backendService.Type()).To(Equal("Backend Service"))
		})
	})
})
