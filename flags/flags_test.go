package flags_test

import (
	"github.com/cloudfoundry/bosh-bootloader/flags"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Flags", func() {
	var (
		f flags.Flags
		// boolVal   bool
		stringVal string
	)

	BeforeEach(func() {
		f = flags.New("test")
		f.String(&stringVal, "string", "")
	})

	Describe("Parse", func() {
		It("can parse strings fields from flags", func() {
			err := f.Parse([]string{"--string", "string_value"})
			Expect(err).NotTo(HaveOccurred())
			Expect(stringVal).To(Equal("string_value"))
		})
	})

	Describe("Args", func() {
		It("returns the remainder of unparsed arguments", func() {
			err := f.Parse([]string{"some-command", "--some-flag"})
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Args()).To(Equal([]string{"some-command", "--some-flag"}))
		})
	})
})
