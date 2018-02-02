package s3_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/genevievelesperance/leftovers/aws/s3"
	"github.com/genevievelesperance/leftovers/aws/s3/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Buckets", func() {
	var (
		client  *fakes.BucketsClient
		logger  *fakes.Logger
		manager *fakes.BucketManager

		buckets s3.Buckets
	)

	BeforeEach(func() {
		client = &fakes.BucketsClient{}
		logger = &fakes.Logger{}
		manager = &fakes.BucketManager{}

		buckets = s3.NewBuckets(client, logger, manager)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptCall.Returns.Proceed = true
			client.ListBucketsCall.Returns.Output = &awss3.ListBucketsOutput{
				Buckets: []*awss3.Bucket{{
					Name: aws.String("banana"),
				}},
			}
			manager.IsInRegionCall.Returns.Output = true
			filter = "ban"
		})

		It("returns a list of s3 buckets to delete", func() {
			items, err := buckets.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListBucketsCall.CallCount).To(Equal(1))
			Expect(manager.IsInRegionCall.CallCount).To(Equal(1))
			Expect(manager.IsInRegionCall.Receives.Bucket).To(Equal("banana"))

			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete bucket banana?"))

			Expect(items).To(HaveLen(1))
			// Expect(items).To(HaveKeyWithValue("banana", ""))
		})

		Context("when the client fails to list buckets", func() {
			BeforeEach(func() {
				client.ListBucketsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error and does not try deleting them", func() {
				_, err := buckets.List(filter)
				Expect(err).To(MatchError("Listing buckets: some error"))
			})
		})

		Context("when the bucket name does not contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := buckets.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(manager.IsInRegionCall.CallCount).To(Equal(0))
				Expect(logger.PromptCall.CallCount).To(Equal(0))

				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the bucket isn't in the region configured", func() {
			BeforeEach(func() {
				manager.IsInRegionCall.Returns.Output = false
			})

			It("does not return it in the list", func() {
				items, err := buckets.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptCall.Returns.Proceed = false
			})

			It("does not delete the bucket", func() {
				items, err := buckets.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete bucket banana?"))

				Expect(items).To(HaveLen(0))
			})
		})
	})
})
