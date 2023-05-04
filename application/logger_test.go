package application_test

import (
	"bytes"
	"fmt"
	"math/rand"

	"github.com/cloudfoundry/bosh-bootloader/application"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger", func() {
	var (
		writer *bytes.Buffer
		reader *bytes.Buffer

		logger *application.Logger
	)

	BeforeEach(func() {
		writer = bytes.NewBuffer([]byte{})
		reader = bytes.NewBuffer([]byte{})

		logger = application.NewLogger(writer, reader)
	})

	Describe("Step", func() {
		It("prints the step message", func() {
			logger.Step("creating key")

			Expect(writer.String()).To(Equal("step: creating key\n"))
		})

		It("prints the step message with dynamic values", func() {
			randomInt := rand.Int()
			logger.Step("Random variable is: %d", randomInt)
			Expect(writer.String()).To(Equal(fmt.Sprintf("step: Random variable is: %d\n", randomInt)))
		})
	})

	Describe("Dot", func() {
		It("prints a dot", func() {
			logger.Dot()
			logger.Dot()
			logger.Dot()

			Expect(writer.String()).To(Equal("\u2022\u2022\u2022"))
		})
	})

	Describe("Println", func() {
		It("prints out the message", func() {
			logger.Println("hello world")

			Expect(writer.String()).To(Equal("hello world\n"))
		})
	})

	Describe("Prompt", func() {
		Context("when NoConfirm has been called", func() {
			BeforeEach(func() {
				logger.NoConfirm()
			})

			It("doesn't prompt", func() {
				proceed := logger.Prompt("do you like cheese?")
				Expect(proceed).To(BeTrue())

				Expect(writer.String()).To(Equal(""))
			})
		})

		It("prompts for the given messge", func() {
			logger.Prompt("do you like cheese?")

			Expect(writer.String()).To(Equal("do you like cheese? (y/N): "))
		})

		DescribeTable("prompting the user for confirmation",
			func(response string, proceed bool) {
				fmt.Fprintf(reader, "%s\n", response)

				p := logger.Prompt("Do you like bananas?")
				Expect(p).To(Equal(proceed))
			},
			Entry("responding with 'yes'", "yes", true),
			Entry("responding with 'y'", "y", true),
			Entry("responding with 'Yes'", "Yes", true),
			Entry("responding with 'Y'", "Y", true),
			Entry("responding with 'no'", "no", false),
			Entry("responding with 'n'", "n", false),
			Entry("responding with 'No'", "No", false),
			Entry("responding with 'N'", "N", false),
		)
	})

	Describe("mixing steps, dots and printlns", func() {
		It("prints out a coherent set of lines", func() {
			logger.Step("creating key")
			logger.Step("generating template")
			logger.Step("applying template")
			logger.Dot()
			logger.Dot()
			logger.Step("completed applying template")
			logger.Dot()
			logger.Dot()
			logger.Prompt("do you like turtles?")
			logger.Println("**bosh manifest**")
			logger.Step("doing more stuff")
			logger.Dot()
			logger.Dot()
			logger.Println("SUCCESS!")

			Expect(writer.String()).To(Equal(`step: creating key
step: generating template
step: applying template
••
step: completed applying template
••
do you like turtles? (y/N): **bosh manifest**
step: doing more stuff
••
SUCCESS!
`))
		})
	})
})
