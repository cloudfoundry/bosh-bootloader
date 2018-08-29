package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("Disks", func() {
	var (
		client *fakes.DisksClient
		logger *fakes.Logger
		zones  map[string]string

		disks compute.Disks
	)

	BeforeEach(func() {
		client = &fakes.DisksClient{}
		logger = &fakes.Logger{}
		zones = map[string]string{"https://zone-1": "zone-1"}

		disks = compute.NewDisks(client, logger, zones)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			filter = "banana"
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListDisksCall.Returns.Output = []*gcpcompute.Disk{{
				Name: "banana-disk",
				Zone: "https://zone-1",
			}, {
				Name: "just-another-disk",
				Zone: "https://zone-2",
			}}
		})

		It("lists, filters, and prompts for disks to delete", func() {
			list, err := disks.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListDisksCall.CallCount).To(Equal(1))
			Expect(client.ListDisksCall.Receives.Zone).To(Equal("zone-1"))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Disk"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-disk"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list disks", func() {
			BeforeEach(func() {
				client.ListDisksCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := disks.List(filter)
				Expect(err).To(MatchError("List Disks for zone zone-1: some error"))
			})
		})

		Context("when the disk name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := disks.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.ListDisksCall.CallCount).To(Equal(1))
				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))

				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the user says no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not add it to the list", func() {
				list, err := disks.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
