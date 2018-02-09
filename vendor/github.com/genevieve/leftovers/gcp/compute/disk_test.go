package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Disk", func() {
	var (
		client *fakes.DisksClient
		name   string
		zone   string

		disk compute.Disk
	)

	BeforeEach(func() {
		client = &fakes.DisksClient{}
		name = "banana"
		zone = "zone"

		disk = compute.NewDisk(client, name, zone)
	})

	Describe("Delete", func() {
		It("deletes the disk", func() {
			err := disk.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteDiskCall.CallCount).To(Equal(1))
			Expect(client.DeleteDiskCall.Receives.Disk).To(Equal(name))
			Expect(client.DeleteDiskCall.Receives.Zone).To(Equal(zone))
		})

		Context("when the client fails to delete", func() {
			BeforeEach(func() {
				client.DeleteDiskCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := disk.Delete()
				Expect(err).To(MatchError("ERROR deleting disk banana: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(disk.Name()).To(Equal(name))
		})
	})
})
