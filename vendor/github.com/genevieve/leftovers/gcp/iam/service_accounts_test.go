package iam_test

import (
	"errors"

	gcpiam "google.golang.org/api/iam/v1"

	"github.com/genevieve/leftovers/gcp/iam"
	"github.com/genevieve/leftovers/gcp/iam/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceAccounts", func() {
	var (
		client *fakes.ServiceAccountsClient
		logger *fakes.Logger

		serviceAccounts iam.ServiceAccounts
	)

	BeforeEach(func() {
		client = &fakes.ServiceAccountsClient{}
		logger = &fakes.Logger{}

		logger.PromptWithDetailsCall.Returns.Proceed = true

		serviceAccounts = iam.NewServiceAccounts(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			client.ListServiceAccountsCall.Returns.Output = []*gcpiam.ServiceAccount{{
				Name: "banana-service-account",
			}}
			filter = "banana"
		})

		It("lists, filters, and prompts for service accounts to delete", func() {
			list, err := serviceAccounts.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListServiceAccountsCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("IAM Service Account"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-service-account"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list serviceAccounts", func() {
			BeforeEach(func() {
				client.ListServiceAccountsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := serviceAccounts.List(filter)
				Expect(err).To(MatchError("List IAM Service Accounts: some error"))
			})
		})

		Context("when the serviceAccount name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := serviceAccounts.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the user says no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not add it to the list", func() {
				list, err := serviceAccounts.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
