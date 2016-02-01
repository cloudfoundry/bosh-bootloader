package application_test

import (
	"bytes"

	"github.com/pivotal-cf-experimental/bosh-bootloader/application"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("App", func() {
	Describe("Run", func() {
		var (
			app    application.App
			stdout *bytes.Buffer
		)

		BeforeEach(func() {
			app = application.New()
			stdout = bytes.NewBuffer([]byte{})
		})

		It("prints out the usage when provided the -h flag", func() {
			Expect(app.Run([]string{"bbl", "-h"}, stdout)).To(Succeed())
			Expect(stdout.String()).To(ContainSubstring("Usage"))
		})

		It("prints out the current version when provided the -v flag", func() {
			Expect(app.Run([]string{"bbl", "-v"}, stdout)).To(Succeed())
			Expect(stdout.String()).To(ContainSubstring("bbl 0.0.1"))
		})

		It("prints an error when an unknown flag is provided", func() {
			err := app.Run([]string{"bbl", "--some-unknown-flag"}, nil)
			Expect(err).To(MatchError("unknown flag `some-unknown-flag'"))
		})
	})
})
