package elbv2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/genevieve/leftovers/aws/elbv2"
	"github.com/genevieve/leftovers/aws/elbv2/fakes"

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
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.DescribeTargetGroupsCall.Returns.Output = &awselbv2.DescribeTargetGroupsOutput{
				TargetGroups: []*awselbv2.TargetGroup{{
					TargetGroupName: aws.String("precursor-banana"),
					TargetGroupArn:  aws.String("precursor-arn"),
				}},
			}
			filter = "banana"
		})

		It("returns a list of target groups to delete", func() {
			items, err := targetGroups.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeTargetGroupsCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("ELBV2 Target Group"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("precursor-banana"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the client fails to describe target groups", func() {
			BeforeEach(func() {
				client.DescribeTargetGroupsCall.Returns.Error = errors.New("error")
			})

			It("returns the error", func() {
				_, err := targetGroups.List(filter)
				Expect(err).To(MatchError("Describe ELBV2 Target Groups: error"))
			})
		})

		Context("when the target group name does not contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := targetGroups.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the user doesn't want to delete", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not return it in the list", func() {
				items, err := targetGroups.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
