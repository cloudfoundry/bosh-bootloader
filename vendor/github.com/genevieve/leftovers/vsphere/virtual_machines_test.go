package vsphere_test

import (
	"errors"

	"github.com/genevieve/leftovers/vsphere"
	"github.com/genevieve/leftovers/vsphere/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Virtual Machines", func() {
	var (
		client *fakes.Client
		logger *fakes.Logger

		virtualMachines vsphere.VirtualMachines
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		client = &fakes.Client{}

		virtualMachines = vsphere.NewVirtualMachines(client, logger)
	})

	Describe("List", func() {
		// The folder requires a valid client for Children()
		PIt("gets the root folder", func() {
			_, err := virtualMachines.List("banana")
			Expect(err).NotTo(HaveOccurred())

			Expect(client.GetRootFolderCall.CallCount).To(Equal(1))
		})

		Context("when getting the root folder fails", func() {
			BeforeEach(func() {
				client.GetRootFolderCall.Returns.Error = errors.New("ruhroh")
			})

			It("returns a helpful error", func() {
				_, err := virtualMachines.List("banana")
				Expect(err).To(MatchError("Getting root folder: ruhroh"))
			})
		})
	})
})
