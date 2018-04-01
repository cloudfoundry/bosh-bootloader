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

var _ = Describe("Key", func() {
	var (
		key      kms.Key
		client   *fakes.KeysClient
		id       *string
		metadata *awskms.KeyMetadata
		tags     []*awskms.Tag
	)

	BeforeEach(func() {
		client = &fakes.KeysClient{}
		id = aws.String("the-id")
		metadata = &awskms.KeyMetadata{Description: aws.String("")}
		tags = []*awskms.Tag{}

		key = kms.NewKey(client, id, metadata, tags)
	})

	Describe("Delete", func() {
		It("deletes the key", func() {
			err := key.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DisableKeyCall.CallCount).To(Equal(1))
			Expect(client.DisableKeyCall.Receives.Input.KeyId).To(Equal(id))
			Expect(client.ScheduleKeyDeletionCall.CallCount).To(Equal(1))
			Expect(client.ScheduleKeyDeletionCall.Receives.Input.KeyId).To(Equal(id))
		})

		Context("when the client fails to disable the key", func() {
			BeforeEach(func() {
				client.DisableKeyCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := key.Delete()
				Expect(err).To(MatchError("Disable: banana"))
			})
		})

		Context("when the client fails to schedule deletion of the key", func() {
			BeforeEach(func() {
				client.ScheduleKeyDeletionCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := key.Delete()
				Expect(err).To(MatchError("Schedule deletion: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(key.Name()).To(Equal("the-id"))
		})

		Context("when there is metadata", func() {
			BeforeEach(func() {
				metadata = &awskms.KeyMetadata{Description: aws.String("bucket key")}
				key = kms.NewKey(client, id, metadata, tags)
			})
			It("returns a more verbose identifier", func() {
				Expect(key.Name()).To(Equal("the-id (Description:bucket key)"))
			})
		})

		Context("when there are tags", func() {
			BeforeEach(func() {
				tags = []*awskms.Tag{{TagKey: aws.String("Key"), TagValue: aws.String("Value")}}
				key = kms.NewKey(client, id, metadata, tags)
			})
			It("returns a more verbose identifier", func() {
				Expect(key.Name()).To(Equal("the-id (Key:Value)"))
			})
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(key.Type()).To(Equal("KMS Key"))
		})
	})
})
