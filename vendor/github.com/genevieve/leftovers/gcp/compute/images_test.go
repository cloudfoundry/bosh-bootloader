package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("Images", func() {
	var (
		client *fakes.ImagesClient
		logger *fakes.Logger

		images compute.Images
	)

	BeforeEach(func() {
		client = &fakes.ImagesClient{}
		logger = &fakes.Logger{}

		images = compute.NewImages(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListImagesCall.Returns.Output = []*gcpcompute.Image{{
				Name: "banana-image",
			}}
			filter = "banana"
		})

		It("lists, filters, and prompts for images to delete", func() {
			list, err := images.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListImagesCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Image"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-image"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list images", func() {
			BeforeEach(func() {
				client.ListImagesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := images.List(filter)
				Expect(err).To(MatchError("List Images: some error"))
			})
		})

		Context("when the image name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := images.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.ListImagesCall.CallCount).To(Equal(1))
				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))

				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the user says no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not add it to the list", func() {
				list, err := images.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
