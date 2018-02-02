package iam_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/genevievelesperance/leftovers/aws/iam"
	"github.com/genevievelesperance/leftovers/aws/iam/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Policies", func() {
	var (
		client *fakes.PoliciesClient
		logger *fakes.Logger

		policies iam.Policies
	)

	BeforeEach(func() {
		client = &fakes.PoliciesClient{}
		logger = &fakes.Logger{}

		policies = iam.NewPolicies(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptCall.Returns.Proceed = true
			client.ListPoliciesCall.Returns.Output = &awsiam.ListPoliciesOutput{
				Policies: []*awsiam.Policy{{
					Arn:        aws.String("the-policy-arn"),
					PolicyName: aws.String("banana-policy"),
				}},
			}
			filter = "banana"
		})

		It("deletes iam policies and associated policies", func() {
			items, err := policies.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListPoliciesCall.CallCount).To(Equal(1))

			Expect(logger.PromptCall.CallCount).To(Equal(1))
			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete policy banana-policy?"))

			Expect(items).To(HaveLen(1))
			// Expect(items).To(HaveKeyWithValue("banana-policy", "the-policy-arn"))
		})

		Context("when the client fails to list policies", func() {
			BeforeEach(func() {
				client.ListPoliciesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error and does not try deleting them", func() {
				_, err := policies.List(filter)
				Expect(err).To(MatchError("Listing policies: some error"))

				Expect(logger.PromptCall.CallCount).To(Equal(0))
			})
		})

		Context("when the policy name does not contain the filter", func() {
			It("does not try to delete it", func() {
				items, err := policies.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptCall.Returns.Proceed = false
			})

			It("does not return it in the list", func() {
				items, err := policies.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete policy banana-policy?"))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
