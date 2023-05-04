package helpers_test

import (
	"crypto/rand"
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("EnvIDGenerator", func() {
	Describe("Generate", func() {
		It("generates a env id with a lake and timestamp", func() {
			generator := helpers.NewEnvIDGenerator(rand.Reader)

			envID, err := generator.Generate()
			Expect(err).NotTo(HaveOccurred())
			Expect(envID).To(MatchRegexp(`bbl-env-([a-z]+-{1}){1,2}\d{4}-\d{2}-\d{2}t\d{2}-\d{2}z`))
		})

		Context("when there are errors", func() {
			It("it returns the error", func() {
				anError := errors.New("banana")
				badReader := fakes.Reader{}
				badReader.ReadCall.Returns.Error = anError

				generator := helpers.NewEnvIDGenerator(&badReader)

				_, err := generator.Generate()
				Expect(err).To(Equal(anError))
			})
		})
	})
})
