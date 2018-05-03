package helpers_test

import (
	"crypto/rand"

	"github.com/cloudfoundry/bosh-bootloader/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PathFinder", func() {
	var pathFinder helpers.PathFinder
	BeforeEach(func() {
		pathFinder = helpers.NewPathFinder()
	})

	Describe("CommandExists", func() {
		var command string
		Context("when a command does not exist", func() {
			BeforeEach(func() {
				commandBytes := make([]byte, 32)

				_, err := rand.Read(commandBytes)
				Expect(err).NotTo(HaveOccurred())

				command = string(commandBytes)
			})

			It("returns false", func() {
				Expect(pathFinder.CommandExists(command)).To(BeFalse())
			})
		})

		Context("when a command exists", func() {
			BeforeEach(func() {
				command = "ginkgo"
			})

			It("returns true", func() {
				Expect(pathFinder.CommandExists(command)).To(BeTrue())
			})
		})
	})
})
