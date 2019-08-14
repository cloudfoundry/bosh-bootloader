package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Router", func() {
	var (
		client *fakes.RoutersClient
		name   string
		region string

		router compute.Router
	)

	BeforeEach(func() {
		client = &fakes.RoutersClient{}
		name = "banana"
		region = "region-1"

		router = compute.NewRouter(client, name, region)
	})

	Describe("Delete", func() {
		It("deletes the router", func() {
			err := router.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteRouterCall.CallCount).To(Equal(1))
			Expect(client.DeleteRouterCall.Receives.Region).To(Equal(region))
			Expect(client.DeleteRouterCall.Receives.Router).To(Equal(name))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteRouterCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := router.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(router.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(router.Type()).To(Equal("Router"))
		})
	})
})
