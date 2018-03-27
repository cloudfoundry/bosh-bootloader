package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/ec2/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Volumes", func() {
	var (
		client *fakes.VolumesClient
		logger *fakes.Logger

		volumes ec2.Volumes
	)

	BeforeEach(func() {
		client = &fakes.VolumesClient{}
		logger = &fakes.Logger{}

		volumes = ec2.NewVolumes(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.DescribeVolumesCall.Returns.Output = &awsec2.DescribeVolumesOutput{
				Volumes: []*awsec2.Volume{{
					VolumeId: aws.String("banana"),
					State:    aws.String("available"),
				}},
			}
			filter = "banana"
		})

		It("deletes ec2 volumes", func() {
			items, err := volumes.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeVolumesCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("volume"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana"))

			Expect(items).To(HaveLen(1))
		})

		PContext("when the volume name does not contain the filter", func() {
			//Volumes do not have names/tags from the environment
		})

		Context("when the client fails to list volumes", func() {
			BeforeEach(func() {
				client.DescribeVolumesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := volumes.List(filter)
				Expect(err).To(MatchError("Describing volumes: some error"))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not delete the volume", func() {
				items, err := volumes.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the volume is not available", func() {
			BeforeEach(func() {
				client.DescribeVolumesCall.Returns.Output = &awsec2.DescribeVolumesOutput{
					Volumes: []*awsec2.Volume{{
						VolumeId: aws.String("banana"),
						State:    aws.String("nope"),
					}},
				}
			})

			It("does not prompt the user and it does not return it in the list", func() {
				items, err := volumes.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
