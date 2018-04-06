package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/ec2/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Snapshot", func() {
	var (
		snapshot ec2.Snapshot
		client   *fakes.SnapshotsClient
		id       *string
	)

	BeforeEach(func() {
		client = &fakes.SnapshotsClient{}
		id = aws.String("the-id")

		snapshot = ec2.NewSnapshot(client, id)
	})

	Describe("Delete", func() {
		It("terminates the snapshot", func() {
			err := snapshot.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteSnapshotCall.CallCount).To(Equal(1))
			Expect(client.DeleteSnapshotCall.Receives.Input.SnapshotId).To(Equal(id))
		})

		Context("when the client fails", func() {
			BeforeEach(func() {
				client.DeleteSnapshotCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := snapshot.Delete()
				Expect(err).To(MatchError("Delete: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(snapshot.Name()).To(Equal("the-id"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(snapshot.Type()).To(Equal("EC2 Snapshot"))
		})
	})
})
