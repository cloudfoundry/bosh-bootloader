package terraform_test

import (
	"bytes"

	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cmd", func() {
	var (
		stdout *bytes.Buffer
		stderr *bytes.Buffer

		cmd terraform.Cmd
	)

	BeforeEach(func() {
		stdout = bytes.NewBuffer([]byte{})
		stderr = bytes.NewBuffer([]byte{})

		cmd = terraform.NewCmd(stderr)
	})

	It("runs terraform with args", func() {
		err := cmd.Run(stdout, "/tmp", []string{"apply", "some-arg"})
		Expect(err).NotTo(HaveOccurred())

		Expect(stdout).To(MatchRegexp("working directory: (.*)/tmp"))
		Expect(stdout).To(ContainSubstring("apply some-arg"))
	})

	Context("failure case", func() {
		It("returns an error when terraform fails", func() {
			err := cmd.Run(stdout, "", []string{"fast-fail"})
			Expect(err).To(MatchError("exit status 1"))

			Expect(stderr).To(ContainSubstring("failed to terraform"))
		})
	})
})
