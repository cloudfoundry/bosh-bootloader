package iam_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/genevieve/leftovers/aws/iam"
	"github.com/genevieve/leftovers/aws/iam/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstanceProfiles", func() {
	var (
		client *fakes.InstanceProfilesClient
		logger *fakes.Logger
		filter string

		instanceProfiles iam.InstanceProfiles
	)

	BeforeEach(func() {
		client = &fakes.InstanceProfilesClient{}
		logger = &fakes.Logger{}
		filter = "banana"

		instanceProfiles = iam.NewInstanceProfiles(client, logger)
	})

	Describe("List", func() {
		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListInstanceProfilesCall.Returns.Output = &awsiam.ListInstanceProfilesOutput{
				InstanceProfiles: []*awsiam.InstanceProfile{{
					InstanceProfileName: aws.String("banana-profile"),
				}},
			}
		})

		It("returns a list of instance profiles to delete", func() {
			items, err := instanceProfiles.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListInstanceProfilesCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("IAM Instance Profile"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-profile"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the instance profile name does not contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := instanceProfiles.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))

				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the client fails to list instance profiles", func() {
			BeforeEach(func() {
				client.ListInstanceProfilesCall.Returns.Error = errors.New("listing error")
			})

			It("returns the error and does not try deleting them", func() {
				_, err := instanceProfiles.List(filter)
				Expect(err).To(MatchError("List IAM Instance Profiles: listing error"))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not return it in the list", func() {
				items, err := instanceProfiles.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))

				Expect(items).To(HaveLen(0))
			})
		})
	})
})
