package boshinit_test

import (
	"bytes"
	"os/exec"

	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandBuilder", func() {
	Describe("DeployCommand", func() {
		It("builds a command with the correct values", func() {
			stdout := bytes.NewBuffer([]byte{})
			stderr := bytes.NewBuffer([]byte{})
			builder := boshinit.NewCommandBuilder("/tmp/bosh-init", "/tmp/some-dir", stdout, stderr)

			cmd := builder.DeployCommand()
			Expect(cmd).To(Equal(&exec.Cmd{
				Path: "/tmp/bosh-init",
				Args: []string{
					"bosh-init",
					"deploy",
					"bosh.yml",
				},
				Dir:    "/tmp/some-dir",
				Stdout: stdout,
				Stderr: stderr,
			}))
		})
	})
})
