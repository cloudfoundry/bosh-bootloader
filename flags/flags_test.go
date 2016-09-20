package flags_test

import (
	"github.com/cloudfoundry/bosh-bootloader/flags"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Flags", func() {
	var (
		f         flags.Flags
		boolVal   bool
		stringVal string
	)

	BeforeEach(func() {
		f = flags.New("test")
		f.Bool(&boolVal, "b", "bool", false)
		f.String(&stringVal, "string", "")
	})

	Describe("Parse", func() {
		Context("Bool flags", func() {
			It("can parse long flags", func() {
				err := f.Parse([]string{"--bool"})
				Expect(err).NotTo(HaveOccurred())
				Expect(boolVal).To(BeTrue())
			})

			It("can parse long flags", func() {
				err := f.Parse([]string{"-b"})
				Expect(err).NotTo(HaveOccurred())
				Expect(boolVal).To(BeTrue())
			})
		})

		Context("String flags", func() {
			It("can parse strings fields from flags", func() {
				err := f.Parse([]string{"--string", "string_value"})
				Expect(err).NotTo(HaveOccurred())
				Expect(stringVal).To(Equal("string_value"))
			})
		})
	})

	Describe("Args", func() {
		It("returns the remainder of unparsed arguments", func() {
			err := f.Parse([]string{"-b", "some-command", "--some-flag"})
			Expect(err).NotTo(HaveOccurred())
			Expect(f.Args()).To(Equal([]string{"some-command", "--some-flag"}))
		})
	})
})
