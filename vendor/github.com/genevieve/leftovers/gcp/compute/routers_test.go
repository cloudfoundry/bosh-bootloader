package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("Routers", func() {
	var (
		client  *fakes.RoutersClient
		logger  *fakes.Logger
		regions map[string]string

		routers compute.Routers
	)

	BeforeEach(func() {
		client = &fakes.RoutersClient{}
		logger = &fakes.Logger{}
		regions = map[string]string{"https://region-1": "region-1"}

		routers = compute.NewRouters(client, logger, regions)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListRoutersCall.Returns.Output = []*gcpcompute.Router{
				{
					Name:   "banana-router",
					Region: "https://region-1",
				},
				{
					Name:   "pineapple-router",
					Region: "https://region-1",
				},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for routers to delete", func() {
			list, err := routers.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListRoutersCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-router"))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Router"))

			Expect(list).To(HaveLen(1))
		})

		Context("when routers client fails to list routers", func() {
			BeforeEach(func() {
				client.ListRoutersCall.Returns.Error = errors.New("some error")
			})

			It("returns helpful error message", func() {
				_, err := routers.List(filter)
				Expect(err).To(MatchError("List Routers for region region-1: some error"))
			})
		})

		Context("when the user does not want to delete resource", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("removes it from the list", func() {
				list, err := routers.List(filter)
				Expect(err).NotTo(HaveOccurred())
				Expect(list).To(HaveLen(0))
			})
		})
	})
})
