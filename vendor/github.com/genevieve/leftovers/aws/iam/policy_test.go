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

var _ = Describe("Policy", func() {
	var (
		policy iam.Policy
		client *fakes.PoliciesClient
		logger *fakes.Logger
		name   *string
		arn    *string
	)

	BeforeEach(func() {
		client = &fakes.PoliciesClient{}
		logger = &fakes.Logger{}
		name = aws.String("banana")
		arn = aws.String("the-arn")

		policy = iam.NewPolicy(client, logger, name, arn)

		client.ListPolicyVersionsCall.Returns.Output = &awsiam.ListPolicyVersionsOutput{
			Versions: []*awsiam.PolicyVersion{},
		}
	})

	Describe("Delete", func() {
		It("deletes the policy", func() {
			err := policy.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeletePolicyCall.CallCount).To(Equal(1))
			Expect(client.DeletePolicyCall.Receives.Input.PolicyArn).To(Equal(arn))
		})

		Context("when the policy has non-default versions", func() {
			BeforeEach(func() {
				client.ListPolicyVersionsCall.Returns.Output = &awsiam.ListPolicyVersionsOutput{
					Versions: []*awsiam.PolicyVersion{
						{IsDefaultVersion: aws.Bool(true), VersionId: aws.String("v2")},
						{IsDefaultVersion: aws.Bool(false), VersionId: aws.String("v1")},
					},
				}
			})

			It("deletes all non-default versions", func() {
				err := policy.Delete()
				Expect(err).NotTo(HaveOccurred())

				Expect(client.ListPolicyVersionsCall.CallCount).To(Equal(1))

				Expect(client.DeletePolicyVersionCall.CallCount).To(Equal(1))
				Expect(client.DeletePolicyVersionCall.Receives.Input.PolicyArn).To(Equal(arn))
				Expect(client.DeletePolicyVersionCall.Receives.Input.VersionId).To(Equal(aws.String("v1")))
			})

			Context("when the client fails to delete policy versions", func() {
				BeforeEach(func() {
					client.DeletePolicyVersionCall.Returns.Error = errors.New("some error")
				})

				It("logs the error", func() {
					err := policy.Delete()
					Expect(err).NotTo(HaveOccurred())

					Expect(logger.PrintfCall.Messages).To(Equal([]string{
						"[IAM Policy: banana] Delete policy version v1: some error \n",
					}))
				})
			})
		})

		Context("when the client fails to delete the policy", func() {
			BeforeEach(func() {
				client.DeletePolicyCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				err := policy.Delete()
				Expect(err).To(MatchError("Delete: some error"))
			})
		})

		Context("when the client fails to list policy versions", func() {
			BeforeEach(func() {
				client.ListPolicyVersionsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				err := policy.Delete()
				Expect(err).To(MatchError("List IAM Policy Versions: some error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(policy.Name()).To(Equal("banana"))
		})
	})

	Describe("Type", func() {
		It("returns \"policy\"", func() {
			Expect(policy.Type()).To(Equal("IAM Policy"))
		})
	})
})
