package main_test

import (
	"bytes"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const COMMAND_TIMEOUT = "1s"

var _ = Describe("bbl", func() {
	Describe("bbl -h", func() {
		It("prints out the usage", func() {
			command := exec.Command(pathToBBL, "-h")
			output := bytes.NewBuffer([]byte{})
			command.Stdout = output

			Eventually(command.Run, COMMAND_TIMEOUT, COMMAND_TIMEOUT).Should(Succeed())
			Expect(output).To(ContainSubstring("Usage"))
		})
	})

	Describe("bbl -v", func() {
		It("prints out the current version", func() {
			command := exec.Command(pathToBBL, "-v")
			output := bytes.NewBuffer([]byte{})
			command.Stdout = output

			Eventually(command.Run, COMMAND_TIMEOUT, COMMAND_TIMEOUT).Should(Succeed())
			Expect(output).To(ContainSubstring("bbl 0.0.1"))
		})
	})

	It("prints an error when an unknown flag is provided", func() {
		command := exec.Command(pathToBBL, "--some-unknown-flag")
		errors := bytes.NewBuffer([]byte{})
		command.Stderr = errors

		Eventually(command.Run, COMMAND_TIMEOUT, COMMAND_TIMEOUT).ShouldNot(Succeed())
		Expect(errors).To(ContainSubstring("unknown flag `some-unknown-flag'"))
	})
})
