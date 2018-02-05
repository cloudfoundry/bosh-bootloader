package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/ec2/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KeyPairs", func() {
	var (
		client *fakes.KeyPairsClient
		logger *fakes.Logger

		keys ec2.KeyPairs
	)

	BeforeEach(func() {
		client = &fakes.KeyPairsClient{}
		logger = &fakes.Logger{}

		keys = ec2.NewKeyPairs(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptCall.Returns.Proceed = true
			client.DescribeKeyPairsCall.Returns.Output = &awsec2.DescribeKeyPairsOutput{
				KeyPairs: []*awsec2.KeyPairInfo{{
					KeyName: aws.String("banana"),
				}},
			}
			filter = "ban"
		})

		It("returns a list of ec2 key pairs to delete", func() {
			items, err := keys.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeKeyPairsCall.CallCount).To(Equal(1))

			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete key pair banana?"))

			Expect(items).To(HaveLen(1))
			// Expect(items).To(HaveKeyWithValue("banana", ""))
		})

		Context("when the client fails to list key pairs", func() {
			BeforeEach(func() {
				client.DescribeKeyPairsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := keys.List(filter)
				Expect(err).To(MatchError("Describing key pairs: some error"))
			})
		})

		Context("when the key pair name does not contain the filter", func() {
			It("does not try deleting it", func() {
				items, err := keys.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptCall.Returns.Proceed = false
			})

			It("does not delete the key pair", func() {
				items, err := keys.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete key pair banana?"))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
