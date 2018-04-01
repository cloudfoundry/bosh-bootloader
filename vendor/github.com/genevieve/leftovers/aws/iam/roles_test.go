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

var _ = Describe("Roles", func() {
	var (
		client   *fakes.RolesClient
		logger   *fakes.Logger
		policies *fakes.RolePolicies
		filter   string

		roles iam.Roles
	)

	BeforeEach(func() {
		client = &fakes.RolesClient{}
		logger = &fakes.Logger{}
		policies = &fakes.RolePolicies{}
		filter = "banana"

		roles = iam.NewRoles(client, logger, policies)
	})

	Describe("List", func() {
		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListRolesCall.Returns.Output = &awsiam.ListRolesOutput{
				Roles: []*awsiam.Role{{
					RoleName: aws.String("banana-role"),
				}},
			}
		})

		It("returns a list of iam roles and associated policies to delete", func() {
			items, err := roles.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListRolesCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("IAM Role"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-role"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the client fails to list roles", func() {
			BeforeEach(func() {
				client.ListRolesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := roles.List(filter)
				Expect(err).To(MatchError("List IAM Roles: some error"))
			})
		})

		Context("when the role name does not contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := roles.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not return it in the list", func() {
				items, err := roles.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
