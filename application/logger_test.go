package application_test

import (
	"bytes"
	"fmt"
	"math/rand"

	"github.com/pivotal-cf-experimental/bosh-bootloader/application"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger", func() {
	var (
		buffer *bytes.Buffer
		logger *application.Logger
	)

	BeforeEach(func() {
		buffer = bytes.NewBuffer([]byte{})
		logger = application.NewLogger(buffer)
	})

	Describe("Step", func() {
		It("prints the step message", func() {
			logger.Step("creating key pair")

			Expect(buffer.String()).To(Equal("step: creating key pair\n"))
		})

		It("prints the step message with dynamic values", func() {
			randomInt := rand.Int()
			logger.Step("Random variable is: %d", randomInt)
			Expect(buffer.String()).To(Equal(fmt.Sprintf("step: Random variable is: %d\n", randomInt)))
		})
	})

	Describe("Dot", func() {
		It("prints a dot", func() {
			logger.Dot()
			logger.Dot()
			logger.Dot()

			Expect(buffer.String()).To(Equal("\u2022\u2022\u2022"))
		})
	})

	Describe("Println", func() {
		It("prints out the message", func() {
			logger.Println("hello world")

			Expect(buffer.String()).To(Equal("hello world\n"))
		})
	})

	Describe("Prompt", func() {
		It("prompts for the given messge", func() {
			logger.Prompt("do you like cheese?")

			Expect(buffer.String()).To(Equal("do you like cheese? (y/N): "))
		})
	})

	Describe("mixing steps, dots and printlns", func() {
		It("prints out a coherent set of lines", func() {
			logger.Step("creating keypair")
			logger.Step("generating cloudformation template")
			logger.Step("applying cloudformation template")
			logger.Dot()
			logger.Dot()
			logger.Step("completed applying cloudformation template")
			logger.Dot()
			logger.Dot()
			logger.Prompt("do you like turtles?")
			logger.Println("**bosh manifest**")
			logger.Step("doing more stuff")
			logger.Dot()
			logger.Dot()
			logger.Println("SUCCESS!")

			Expect(buffer.String()).To(Equal(`step: creating keypair
step: generating cloudformation template
step: applying cloudformation template
••
step: completed applying cloudformation template
••
do you like turtles? (y/N): **bosh manifest**
step: doing more stuff
••
SUCCESS!
`))
		})
	})
})
