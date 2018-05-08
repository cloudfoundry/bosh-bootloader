package storage_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/storage"
	"github.com/genevieve/leftovers/gcp/storage/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bucket", func() {
	var (
		client *fakes.BucketsClient
		name   string

		bucket storage.Bucket
	)

	BeforeEach(func() {
		client = &fakes.BucketsClient{}
		name = "banana"

		bucket = storage.NewBucket(client, name)
	})

	Describe("Delete", func() {
		It("deletes the bucket", func() {
			err := bucket.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteBucketCall.CallCount).To(Equal(1))
			Expect(client.DeleteBucketCall.Receives.Bucket).To(Equal(name))
		})

		Context("when the client fails to delete the bucket", func() {
			BeforeEach(func() {
				client.DeleteBucketCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := bucket.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(bucket.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(bucket.Type()).To(Equal("Storage Bucket"))
		})
	})
})
