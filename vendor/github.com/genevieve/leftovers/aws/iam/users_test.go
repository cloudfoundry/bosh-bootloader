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

var _ = Describe("Users", func() {
	var (
		client   *fakes.UsersClient
		logger   *fakes.Logger
		policies *fakes.UserPolicies
		keys     *fakes.AccessKeys
		filter   string

		users iam.Users
	)

	BeforeEach(func() {
		client = &fakes.UsersClient{}
		logger = &fakes.Logger{}
		policies = &fakes.UserPolicies{}
		keys = &fakes.AccessKeys{}
		filter = "banana"

		users = iam.NewUsers(client, logger, policies, keys)
	})

	Describe("List", func() {
		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListUsersCall.Returns.Output = &awsiam.ListUsersOutput{
				Users: []*awsiam.User{{
					UserName: aws.String("banana-user"),
				}},
			}
		})

		It("returns a list of iam users to delete", func() {
			items, err := users.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListUsersCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("IAM User"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-user"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the client fails to list users", func() {
			BeforeEach(func() {
				client.ListUsersCall.Returns.Error = errors.New("some error")
			})

			It("returns the error and does not try deleting them", func() {
				_, err := users.List(filter)
				Expect(err).To(MatchError("List IAM Users: some error"))
			})
		})

		Context("when the user name does not contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := users.List("kiwi")
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
				items, err := users.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
