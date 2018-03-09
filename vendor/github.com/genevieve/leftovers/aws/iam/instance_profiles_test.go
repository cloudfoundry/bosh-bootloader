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

		instanceProfiles iam.InstanceProfiles
	)

	BeforeEach(func() {
		client = &fakes.InstanceProfilesClient{}
		logger = &fakes.Logger{}

		instanceProfiles = iam.NewInstanceProfiles(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptCall.Returns.Proceed = true
			client.ListInstanceProfilesCall.Returns.Output = &awsiam.ListInstanceProfilesOutput{
				InstanceProfiles: []*awsiam.InstanceProfile{{
					InstanceProfileName: aws.String("banana-profile"),
				}},
			}
			filter = "banana"
		})

		It("detaches roles and returns a list of instance profiles to delete", func() {
			items, err := instanceProfiles.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListInstanceProfilesCall.CallCount).To(Equal(1))

			Expect(logger.PromptCall.CallCount).To(Equal(1))
			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete instance profile banana-profile?"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the instance profile name does not contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := instanceProfiles.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.CallCount).To(Equal(0))

				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the client fails to list instance profiles", func() {
			BeforeEach(func() {
				client.ListInstanceProfilesCall.Returns.Error = errors.New("listing error")
			})

			It("returns the error and does not try deleting them", func() {
				_, err := instanceProfiles.List(filter)
				Expect(err).To(MatchError("Listing instance profiles: listing error"))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptCall.Returns.Proceed = false
			})

			It("does not return it in the list", func() {
				items, err := instanceProfiles.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete instance profile banana-profile?"))

				Expect(items).To(HaveLen(0))
			})
		})
	})
})
