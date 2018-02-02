package elbv2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/genevievelesperance/leftovers/aws/elbv2"
	"github.com/genevievelesperance/leftovers/aws/elbv2/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TargetGroups", func() {
	var (
		client *fakes.TargetGroupsClient
		logger *fakes.Logger

		targetGroups elbv2.TargetGroups
	)

	BeforeEach(func() {
		client = &fakes.TargetGroupsClient{}
		logger = &fakes.Logger{}

		targetGroups = elbv2.NewTargetGroups(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptCall.Returns.Proceed = true
			client.DescribeTargetGroupsCall.Returns.Output = &awselbv2.DescribeTargetGroupsOutput{
				TargetGroups: []*awselbv2.TargetGroup{{
					TargetGroupName: aws.String("precursor-banana"),
					TargetGroupArn:  aws.String("precursor-arn"),
				}, {
					TargetGroupName: aws.String("banana"),
					TargetGroupArn:  aws.String("arn"),
				}},
			}
			filter = "banana"
		})

		It("returns a list of target groups to delete", func() {
			items, err := targetGroups.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.PromptCall.CallCount).To(Equal(2))

			Expect(items).To(HaveLen(2))
			// Expect(items).To(HaveKeyWithValue("banana", "arn"))
			// Expect(items).To(HaveKeyWithValue("precursor-banana", "precursor-arn"))
		})

		Context("when the client fails to describe target groups", func() {
			BeforeEach(func() {
				client.DescribeTargetGroupsCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				_, err := targetGroups.List(filter)
				Expect(err).To(MatchError("Describing target groups: banana"))
			})
		})

		Context("when the target group name does not contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := targetGroups.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.CallCount).To(Equal(0))
				Expect(client.DeleteTargetGroupCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the user doesn't want to delete", func() {
			BeforeEach(func() {
				logger.PromptCall.Returns.Proceed = false
			})

			It("does not return it in the list", func() {
				items, err := targetGroups.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.CallCount).To(Equal(2))
				Expect(client.DeleteTargetGroupCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
