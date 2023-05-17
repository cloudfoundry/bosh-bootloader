package helpers_test

import (
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PathFinder", func() {
	var pathFinder helpers.PathFinder
	BeforeEach(func() {
		pathFinder = helpers.NewPathFinder()
	})

	Describe("CommandExists", func() {
		Context("when a command does not exist", func() {
			It("returns false", func() {
				Expect(pathFinder.CommandExists("non-existent-command")).To(BeFalse())
			})
		})

		Context("when a command exists", func() {
			It("returns true", func() {
				Expect(pathFinder.CommandExists("ls")).To(BeTrue())
			})
		})
	})
})
