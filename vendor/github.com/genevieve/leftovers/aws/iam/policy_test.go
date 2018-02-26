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
		name   *string
		arn    *string
	)

	BeforeEach(func() {
		client = &fakes.PoliciesClient{}
		name = aws.String("the-name")
		arn = aws.String("the-arn")

		policy = iam.NewPolicy(client, name, arn)

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
					client.DeletePolicyVersionCall.Returns.Error = errors.New("banana")
				})

				It("returns the error", func() {
					err := policy.Delete()
					Expect(err).To(MatchError("FAILED deleting version v1 of policy the-name: banana"))
				})
			})
		})

		Context("when the client fails to delete the policy", func() {
			BeforeEach(func() {
				client.DeletePolicyCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := policy.Delete()
				Expect(err).To(MatchError("FAILED deleting policy the-name: banana"))
			})
		})

		Context("when the client fails to list policy versions", func() {
			BeforeEach(func() {
				client.ListPolicyVersionsCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := policy.Delete()
				Expect(err).To(MatchError("FAILED listing versions for policy the-name: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(policy.Name()).To(Equal("the-name"))
		})
	})
})
