package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("UrlMaps", func() {
	var (
		client *fakes.UrlMapsClient
		logger *fakes.Logger

		urlMaps compute.UrlMaps
	)

	BeforeEach(func() {
		client = &fakes.UrlMapsClient{}
		logger = &fakes.Logger{}

		urlMaps = compute.NewUrlMaps(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListUrlMapsCall.Returns.Output = &gcpcompute.UrlMapList{
				Items: []*gcpcompute.UrlMap{{
					Name: "banana-url-map",
				}},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for url maps to delete", func() {
			list, err := urlMaps.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListUrlMapsCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Url Map"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-url-map"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list url maps", func() {
			BeforeEach(func() {
				client.ListUrlMapsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := urlMaps.List(filter)
				Expect(err).To(MatchError("List Url Maps: some error"))
			})
		})

		Context("when the url map name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := urlMaps.List("grape")
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
				list, err := urlMaps.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
