package iam_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/genevieve/leftovers/aws/iam"
	"github.com/genevieve/leftovers/aws/iam/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserPolicies", func() {
	var (
		client *fakes.UserPoliciesClient
		logger *fakes.Logger

		policies iam.UserPolicies
	)

	BeforeEach(func() {
		client = &fakes.UserPoliciesClient{}
		logger = &fakes.Logger{}

		policies = iam.NewUserPolicies(client, logger)
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			client.ListAttachedUserPoliciesCall.Returns.Output = &awsiam.ListAttachedUserPoliciesOutput{
				AttachedPolicies: []*awsiam.AttachedPolicy{{
					PolicyName: aws.String("the-policy"),
					PolicyArn:  aws.String("the-policy-arn"),
				}},
			}
		})

		It("detaches and deletes the policies", func() {
			err := policies.Delete("banana")
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListAttachedUserPoliciesCall.CallCount).To(Equal(1))
			Expect(client.ListAttachedUserPoliciesCall.Receives.Input.UserName).To(Equal(aws.String("banana")))

			Expect(client.DetachUserPolicyCall.CallCount).To(Equal(1))
			Expect(client.DetachUserPolicyCall.Receives.Input.UserName).To(Equal(aws.String("banana")))
			Expect(client.DetachUserPolicyCall.Receives.Input.PolicyArn).To(Equal(aws.String("the-policy-arn")))

			Expect(client.DeleteUserPolicyCall.CallCount).To(Equal(1))
			Expect(client.DeleteUserPolicyCall.Receives.Input.UserName).To(Equal(aws.String("banana")))
			Expect(client.DeleteUserPolicyCall.Receives.Input.PolicyName).To(Equal(aws.String("the-policy")))

			Expect(logger.PrintfCall.Messages).To(Equal([]string{
				"[IAM User: banana] Detached policy the-policy \n",
				"[IAM User: banana] Deleted policy the-policy \n",
			}))
		})

		Context("when the client fails to list attached user policies", func() {
			BeforeEach(func() {
				client.ListAttachedUserPoliciesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error and does not try deleting them", func() {
				err := policies.Delete("banana")
				Expect(err).To(MatchError("List IAM User Policies: some error"))

				Expect(client.DetachUserPolicyCall.CallCount).To(Equal(0))
				Expect(client.DeleteUserPolicyCall.CallCount).To(Equal(0))
			})
		})

		Context("when the client fails to detach the user policy", func() {
			BeforeEach(func() {
				client.DetachUserPolicyCall.Returns.Error = errors.New("some error")
			})

			It("logs the error and deletes the user policy", func() {
				err := policies.Delete("banana")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DeleteUserPolicyCall.CallCount).To(Equal(1))
				Expect(logger.PrintfCall.Messages).To(Equal([]string{
					"[IAM User: banana] Detach policy the-policy: some error \n",
					"[IAM User: banana] Deleted policy the-policy \n",
				}))
			})
		})

		Context("when the client fails to detach the user policy due to NoSuchEntity", func() {
			BeforeEach(func() {
				client.DetachUserPolicyCall.Returns.Error = awserr.New("NoSuchEntity", "hi", nil)
			})

			It("logs success", func() {
				err := policies.Delete("banana")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DeleteUserPolicyCall.CallCount).To(Equal(1))
				Expect(logger.PrintfCall.Messages).To(Equal([]string{
					"[IAM User: banana] Detached policy the-policy \n",
					"[IAM User: banana] Deleted policy the-policy \n",
				}))
			})
		})

		Context("when the client fails to delete the user policy", func() {
			BeforeEach(func() {
				client.DeleteUserPolicyCall.Returns.Error = errors.New("some error")
			})

			It("logs the error", func() {
				err := policies.Delete("banana")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PrintfCall.Messages).To(Equal([]string{
					"[IAM User: banana] Detached policy the-policy \n",
					"[IAM User: banana] Delete policy the-policy: some error \n",
				}))
			})
		})

		Context("when the client fails to delete the user policy due to NoSuchEntity", func() {
			BeforeEach(func() {
				client.DetachUserPolicyCall.Returns.Error = awserr.New("NoSuchEntity", "hi", nil)
				client.DeleteUserPolicyCall.Returns.Error = awserr.New("NoSuchEntity", "hi", nil)
			})

			It("logs success", func() {
				err := policies.Delete("banana")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DeleteUserPolicyCall.CallCount).To(Equal(1))
				Expect(logger.PrintfCall.Messages).To(Equal([]string{
					"[IAM User: banana] Detached policy the-policy \n",
					"[IAM User: banana] Deleted policy the-policy \n",
				}))
			})
		})
	})
})
