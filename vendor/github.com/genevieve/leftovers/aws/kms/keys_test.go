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

var _ = Describe("Keys", func() {
	var (
		client *fakes.KeysClient
		logger *fakes.Logger

		keys kms.Keys
	)

	BeforeEach(func() {
		client = &fakes.KeysClient{}
		logger = &fakes.Logger{}

		keys = kms.NewKeys(client, logger)
	})

	Describe("List", func() {
		var filter string
		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListKeysCall.Returns.Output = &awskms.ListKeysOutput{
				Keys: []*awskms.KeyListEntry{{
					KeyId: aws.String("banana"),
				}},
			}
			client.DescribeKeyCall.Returns.Output = &awskms.DescribeKeyOutput{
				KeyMetadata: &awskms.KeyMetadata{
					Description: aws.String(""),
					KeyState:    aws.String("Enabled"),
				},
			}
			client.ListResourceTagsCall.Returns.Output = &awskms.ListResourceTagsOutput{
				Tags: []*awskms.Tag{},
			}
			filter = "ban"
		})

		It("returns a list of kms keys to delete", func() {
			items, err := keys.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListKeysCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("KMS Key"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the alias name does not contain the filter", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = true
				client.ListKeysCall.Returns.Output = &awskms.ListKeysOutput{
					Keys: []*awskms.KeyListEntry{{
						KeyId: aws.String("banana"),
					}},
				}
				filter = "kiwi"
			})

			It("does not return it in the list", func() {
				items, err := keys.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(client.ListKeysCall.CallCount).To(Equal(1))
				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the client fails to list keys", func() {
			BeforeEach(func() {
				client.ListKeysCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := keys.List(filter)
				Expect(err).To(MatchError("Listing KMS Keys: some error"))
			})
		})

		Context("when the client fails to describe a key", func() {
			BeforeEach(func() {
				client.DescribeKeyCall.Returns.Error = errors.New("some error")
			})

			It("ignores the error", func() {
				_, err := keys.List(filter)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the client fails to list resource tags", func() {
			BeforeEach(func() {
				client.ListResourceTagsCall.Returns.Error = errors.New("some error")
			})

			It("ignores the error", func() {
				_, err := keys.List(filter)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not return it in the list", func() {
				items, err := keys.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
