package commands_test

import (
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CleanupLeftovers", func() {
	var (
		filter  string
		deleter *fakes.FilteredDeleter
		cleanup commands.CleanupLeftovers
	)

	BeforeEach(func() {
		filter = "banana"
		deleter = &fakes.FilteredDeleter{}
		cleanup = commands.NewCleanupLeftovers(deleter)
	})

	Describe("Execute", func() {
		It("calls delete on leftovers with the filter", func() {
			err := cleanup.Execute([]string{"--filter", filter}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(deleter.DeleteCall.CallCount).To(Equal(1))
			Expect(deleter.DeleteCall.Receives.Filter).To(Equal(filter))
		})

		Context("on vsphere", func() {
			Context("without a filter", func() {
				It("returns a helpful error message", func() {
					err := cleanup.Execute([]string{}, storage.State{IAAS: "vsphere"})
					Expect(err).To(MatchError(ContainSubstring("cleanup-leftovers on vSphere requires a filter.")))
				})
			})
		})

		Context("with --dry-run", func() {
			It("lists rather than deleting resources", func() {
				err := cleanup.Execute([]string{"--dry-run", "--filter", filter}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(deleter.ListCall.CallCount).To(Equal(1))
				Expect(deleter.ListCall.Receives.Filter).To(Equal(filter))
			})
		})

		Context("when parsing flags throws an error", func() {
			It("returns a helpful message", func() {
				err := cleanup.Execute([]string{"--filter"}, storage.State{})
				Expect(err).To(MatchError(ContainSubstring("Parsing cleanup-leftovers args: flag needs an argument")))

				Expect(deleter.DeleteCall.CallCount).To(Equal(0))
			})
		})
	})
})
