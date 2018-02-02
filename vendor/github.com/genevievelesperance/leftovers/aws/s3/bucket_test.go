package s3_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevievelesperance/leftovers/aws/s3"
	"github.com/genevievelesperance/leftovers/aws/s3/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bucket", func() {
	var (
		bucket s3.Bucket
		client *fakes.BucketsClient
		name   *string
	)

	BeforeEach(func() {
		client = &fakes.BucketsClient{}
		name = aws.String("the-name")

		bucket = s3.NewBucket(client, name)
	})

	It("deletes the bucket", func() {
		err := bucket.Delete()
		Expect(err).NotTo(HaveOccurred())

		Expect(client.DeleteBucketCall.CallCount).To(Equal(1))
		Expect(client.DeleteBucketCall.Receives.Input.Bucket).To(Equal(name))
	})

	Context("the client fails", func() {
		BeforeEach(func() {
			client.DeleteBucketCall.Returns.Error = errors.New("banana")
		})

		It("returns the error", func() {
			err := bucket.Delete()
			Expect(err).To(MatchError("FAILED deleting bucket the-name: banana"))
		})
	})
})
