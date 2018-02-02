package compute_test

import (
	"errors"

	"github.com/genevievelesperance/leftovers/gcp/compute"
	"github.com/genevievelesperance/leftovers/gcp/compute/fakes"
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
			logger.PromptCall.Returns.Proceed = true
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

			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete url map banana-url-map?"))

			Expect(list).To(HaveLen(1))
			Expect(list).To(HaveKeyWithValue("banana-url-map", ""))
		})

		Context("when the client fails to list url maps", func() {
			BeforeEach(func() {
				client.ListUrlMapsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := urlMaps.List(filter)
				Expect(err).To(MatchError("Listing url maps: some error"))
			})
		})

		Context("when the url map name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := urlMaps.List("grape")
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
				list, err := urlMaps.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})

	Describe("Delete", func() {
		var list map[string]string

		BeforeEach(func() {
			list = map[string]string{"banana-url-map": ""}
		})

		It("deletes url maps", func() {
			urlMaps.Delete(list)

			Expect(client.DeleteUrlMapCall.CallCount).To(Equal(1))
			Expect(client.DeleteUrlMapCall.Receives.UrlMap).To(Equal("banana-url-map"))

			Expect(logger.PrintfCall.Messages).To(Equal([]string{"SUCCESS deleting url map banana-url-map\n"}))
		})

		Context("when the client fails to delete a url map", func() {
			BeforeEach(func() {
				client.DeleteUrlMapCall.Returns.Error = errors.New("some error")
			})

			It("logs the error", func() {
				urlMaps.Delete(list)

				Expect(logger.PrintfCall.Messages).To(Equal([]string{"ERROR deleting url map banana-url-map: some error\n"}))
			})
		})
	})
})
