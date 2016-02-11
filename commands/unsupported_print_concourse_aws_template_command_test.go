package commands_test

import (
	"bytes"
	"io/ioutil"

	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnsupportedPrintConcourseAWSTemplateCommand", func() {
	Describe("Execute", func() {
		It("prints a CloudFormation template", func() {
			stdout := bytes.NewBuffer([]byte{})
			command := commands.NewUnsupportedPrintConcourseAWSTemplateCommand(stdout)

			err := command.Execute([]string{})
			Expect(err).NotTo(HaveOccurred())

			buf, err := ioutil.ReadFile("fixtures/cloudformation.json")
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout.String()).To(MatchJSON(string(buf)))
		})
	})
})
