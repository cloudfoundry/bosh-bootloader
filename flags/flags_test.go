package flags_test

import (
	"github.com/cloudfoundry/bosh-bootloader/flags"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Flags", func() {
	var (
		f         flags.Flags
		stringVal string
		boolVal   bool
	)

	BeforeEach(func() {
		f = flags.New("test")
		f.String(&stringVal, "string", "")
		f.Bool(&boolVal, "bool")
	})

	Describe("Parse", func() {
		It("can parse strings fields from flags", func() {
			err := f.Parse([]string{"--string", "string_value"})
			Expect(err).NotTo(HaveOccurred())
			Expect(stringVal).To(Equal("string_value"))
		})

		It("can parse boolean flags", func() {
			err := f.Parse([]string{"--bool"})
			Expect(err).NotTo(HaveOccurred())
			Expect(boolVal).To(BeTrue())
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
