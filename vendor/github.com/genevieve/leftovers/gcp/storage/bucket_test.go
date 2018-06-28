package storage_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/storage"
	"github.com/genevieve/leftovers/gcp/storage/fakes"
	gcpstorage "google.golang.org/api/storage/v1"

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
		BeforeEach(func() {
			client.ListObjectsCall.Returns.Objects = &gcpstorage.Objects{
				Items: []*gcpstorage.Object{{
					Name:       "canteloupe",
					Generation: int64(1),
				}},
			}
		})

		It("lists objects, deletes objects, then deletes the bucket", func() {
			err := bucket.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListObjectsCall.CallCount).To(Equal(1))
			Expect(client.ListObjectsCall.Receives.Bucket).To(Equal(name))

			Expect(client.DeleteObjectCall.CallCount).To(Equal(1))
			Expect(client.DeleteObjectCall.Receives.Bucket).To(Equal(name))
			Expect(client.DeleteObjectCall.Receives.Object).To(Equal("canteloupe"))
			Expect(client.DeleteObjectCall.Receives.Generation).To(Equal(int64(1)))

			Expect(client.DeleteBucketCall.CallCount).To(Equal(1))
			Expect(client.DeleteBucketCall.Receives.Bucket).To(Equal(name))
		})

		Context("when the are no objects in the bucket", func() {
			BeforeEach(func() {
				client.ListObjectsCall.Returns.Objects = &gcpstorage.Objects{
					Items: []*gcpstorage.Object{},
				}
			})

			It("does not try to delete objects", func() {
				_ = bucket.Delete()
				Expect(client.DeleteObjectCall.CallCount).To(Equal(0))
			})
		})

		Context("when the client fails to list objects in the bucket", func() {
			BeforeEach(func() {
				client.ListObjectsCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := bucket.Delete()
				Expect(err).To(MatchError("List Objects: the-error"))
			})
		})

		Context("when the client fails to delete objects in the bucket", func() {
			BeforeEach(func() {
				client.DeleteObjectCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := bucket.Delete()
				Expect(err).To(MatchError("Delete Object: the-error"))
			})
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
