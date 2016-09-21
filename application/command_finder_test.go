package application_test

import (
	"github.com/cloudfoundry/bosh-bootloader/application"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandFinder", func() {

	DescribeTable("FindCommand", func(input []string, expectedOutput application.CommandFinderResult) {
		parserOutput := application.NewCommandFinder().FindCommand(input)
		Expect(parserOutput).To(Equal(expectedOutput))
	},
		Entry("parses the first non-hyphenated word as the attempted command",
			[]string{"--badflag", "x", "help", "delete-lbs", "--other-flag"},
			application.CommandFinderResult{GlobalFlags: []string{"--badflag"}, Command: "x", OtherArgs: []string{"help", "delete-lbs", "--other-flag"}}),
		Entry("parses the first non-hyphenated word as the state-dir if it directly follows state-dir",
			[]string{"--state-dir", "help", "delete-errthing", "--other-flag"},
			application.CommandFinderResult{GlobalFlags: []string{"--state-dir", "help"}, Command: "delete-errthing", OtherArgs: []string{"--other-flag"}}),
		Entry("parses the first non-hyphenated word as the attempted command if --state-dir=x is provided",
			[]string{"--state-dir=some-dir", "help", "--other-flag"},
			application.CommandFinderResult{GlobalFlags: []string{"--state-dir=some-dir"}, Command: "help", OtherArgs: []string{"--other-flag"}}),
		Entry("parses correctly if no global flags given",
			[]string{"help", "foo", "--other-flag"},
			application.CommandFinderResult{GlobalFlags: []string{}, Command: "help", OtherArgs: []string{"foo", "--other-flag"}}),
		Entry("parses correctly if no other flags given",
			[]string{"help"},
			application.CommandFinderResult{GlobalFlags: []string{}, Command: "help", OtherArgs: []string{}}),
	)

})
