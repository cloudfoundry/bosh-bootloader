package storage_test

import (
	"errors"

	gcpstorage "google.golang.org/api/storage/v1"

	"github.com/genevieve/leftovers/gcp/storage"
	"github.com/genevieve/leftovers/gcp/storage/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Buckets", func() {
	var (
		client *fakes.BucketsClient
		logger *fakes.Logger

		buckets storage.Buckets
	)

	BeforeEach(func() {
		client = &fakes.BucketsClient{}
		logger = &fakes.Logger{}

		logger.PromptWithDetailsCall.Returns.Proceed = true

		buckets = storage.NewBuckets(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			client.ListBucketsCall.Returns.Output = &gcpstorage.Buckets{
				Items: []*gcpstorage.Bucket{{
					Name: "banana-bucket",
				}},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for buckets to delete", func() {
			list, err := buckets.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListBucketsCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Storage Bucket"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-bucket"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list buckets", func() {
			BeforeEach(func() {
				client.ListBucketsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := buckets.List(filter)
				Expect(err).To(MatchError("List Storage Buckets: some error"))
			})
		})

		Context("when the bucket name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := buckets.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the bucket says no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not add it to the list", func() {
				list, err := buckets.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
