package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/ec2/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Image", func() {
	var (
		client       *fakes.ImagesClient
		imageId      *string
		resourceTags *fakes.ResourceTags

		address ec2.Image
	)

	BeforeEach(func() {
		client = &fakes.ImagesClient{}
		imageId = aws.String("the-image-id")
		resourceTags = &fakes.ResourceTags{}

		address = ec2.NewImage(client, imageId, resourceTags)
	})

	Describe("Delete", func() {
		It("releases the address", func() {
			err := address.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeregisterImageCall.CallCount).To(Equal(1))
			Expect(client.DeregisterImageCall.Receives.Input.ImageId).To(Equal(imageId))

			Expect(resourceTags.DeleteCall.CallCount).To(Equal(1))
			Expect(resourceTags.DeleteCall.Receives.ResourceType).To(Equal("image"))
			Expect(resourceTags.DeleteCall.Receives.ResourceId).To(Equal("the-image-id"))
		})

		Context("when the client fails", func() {
			BeforeEach(func() {
				client.DeregisterImageCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := address.Delete()
				Expect(err).To(MatchError("Delete: banana"))
			})
		})

		Context("when the resource tags fails", func() {
			BeforeEach(func() {
				resourceTags.DeleteCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := address.Delete()
				Expect(err).To(MatchError("Delete tags: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(address.Name()).To(Equal("the-image-id"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(address.Type()).To(Equal("EC2 Image"))
		})
	})
})
