package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/ec2/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Volume", func() {
	var (
		volume ec2.Volume
		client *fakes.VolumesClient
		id     *string
		state  *string
	)

	BeforeEach(func() {
		client = &fakes.VolumesClient{}
		id = aws.String("the-id")
		state = aws.String("available")
		tags := []*awsec2.Tag{{Key: aws.String("hi"), Value: aws.String("bye")}}

		volume = ec2.NewVolume(client, id, state, tags)
	})

	Describe("Delete", func() {
		It("deletes the volume", func() {
			err := volume.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteVolumeCall.CallCount).To(Equal(1))
			Expect(client.DeleteVolumeCall.Receives.Input.VolumeId).To(Equal(id))
		})

		Context("the volume has already been deleted", func() {
			BeforeEach(func() {
				client.DeleteVolumeCall.Returns.Error = awserr.New("InvalidVolume.NotFound", "msg", nil)
			})

			It("returns nil", func() {
				err := volume.Delete()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("the client fails", func() {
			BeforeEach(func() {
				client.DeleteVolumeCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := volume.Delete()
				Expect(err).To(MatchError("Delete: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(volume.Name()).To(Equal("the-id (State:available) (hi:bye)"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(volume.Type()).To(Equal("EC2 Volume"))
		})
	})
})
