package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	awssts "github.com/aws/aws-sdk-go/service/sts"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/ec2/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Images", func() {
	var (
		client       *fakes.ImagesClient
		stsClient    *fakes.StsClient
		logger       *fakes.Logger
		resourceTags *fakes.ResourceTags

		images ec2.Images
	)

	BeforeEach(func() {
		client = &fakes.ImagesClient{}
		stsClient = &fakes.StsClient{}
		logger = &fakes.Logger{}
		logger.PromptWithDetailsCall.Returns.Proceed = true
		resourceTags = &fakes.ResourceTags{}

		images = ec2.NewImages(client, stsClient, logger, resourceTags)
	})

	Describe("List", func() {
		BeforeEach(func() {
			client.DescribeImagesCall.Returns.Output = &awsec2.DescribeImagesOutput{
				Images: []*awsec2.Image{{
					ImageId: aws.String("the-image-id"),
				}},
			}
			stsClient.GetCallerIdentityCall.Returns.Output = &awssts.GetCallerIdentityOutput{
				Account: aws.String("the-account-id"),
			}
		})

		It("returns a list of ec2 images to delete", func() {
			items, err := images.List("")
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeImagesCall.CallCount).To(Equal(1))
			Expect(client.DescribeImagesCall.Receives.Input.Owners[0]).To(Equal(aws.String("the-account-id")))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("EC2 Image"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("the-image-id"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the client fails to list images", func() {
			BeforeEach(func() {
				client.DescribeImagesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := images.List("")
				Expect(err).To(MatchError("Describing EC2 Images: some error"))
			})
		})

		Context("when the sts client fails to return the caller identity", func() {
			BeforeEach(func() {
				stsClient.GetCallerIdentityCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := images.List("")
				Expect(err).To(MatchError("Get caller identity: some error"))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not return it to the list", func() {
				items, err := images.List("")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
