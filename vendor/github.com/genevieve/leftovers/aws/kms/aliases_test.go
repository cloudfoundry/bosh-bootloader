package kms_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awskms "github.com/aws/aws-sdk-go/service/kms"
	"github.com/genevieve/leftovers/aws/kms"
	"github.com/genevieve/leftovers/aws/kms/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Aliases", func() {
	var (
		client *fakes.AliasesClient
		logger *fakes.Logger

		aliases kms.Aliases
	)

	BeforeEach(func() {
		client = &fakes.AliasesClient{}
		logger = &fakes.Logger{}

		aliases = kms.NewAliases(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListAliasesCall.Returns.Output = &awskms.ListAliasesOutput{
				Aliases: []*awskms.AliasListEntry{{
					AliasName: aws.String("banana"),
				}},
			}
			filter = "ban"
		})

		It("returns a list of kms aliases to delete", func() {
			items, err := aliases.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListAliasesCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("KMS Alias"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the alias name does not contain the filter", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = true
				client.ListAliasesCall.Returns.Output = &awskms.ListAliasesOutput{
					Aliases: []*awskms.AliasListEntry{{
						AliasName: aws.String("nope"),
					}},
				}
				filter = "banana"
			})

			It("does not return it in the list", func() {
				items, err := aliases.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(client.ListAliasesCall.CallCount).To(Equal(1))
				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the client fails to describe aliases", func() {
			BeforeEach(func() {
				client.ListAliasesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := aliases.List(filter)
				Expect(err).To(MatchError("Listing KMS Aliases: some error"))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not return it in the list", func() {
				items, err := aliases.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
