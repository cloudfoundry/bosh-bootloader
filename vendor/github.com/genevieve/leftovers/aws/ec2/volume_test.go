package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
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
	)

	BeforeEach(func() {
		client = &fakes.VolumesClient{}
		id = aws.String("the-id")

		volume = ec2.NewVolume(client, id)
	})

	Describe("Delete", func() {
		It("deletes the volume", func() {
			err := volume.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteVolumeCall.CallCount).To(Equal(1))
			Expect(client.DeleteVolumeCall.Receives.Input.VolumeId).To(Equal(id))
		})

		Context("the client fails", func() {
			BeforeEach(func() {
				client.DeleteVolumeCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := volume.Delete()
				Expect(err).To(MatchError("FAILED deleting volume the-id: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(volume.Name()).To(Equal("the-id"))
		})
	})
})
