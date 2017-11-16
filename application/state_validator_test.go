package application_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StateValidator", func() {
	var (
		tempDirectory  string
		stateValidator application.StateValidator
	)

	BeforeEach(func() {
		var err error
		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		stateValidator = application.NewStateValidator(tempDirectory)
	})

	Context("when state file exists", func() {
		BeforeEach(func() {
			err := ioutil.WriteFile(filepath.Join(tempDirectory, "bbl-state.json"), []byte(""), storage.StateMode)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns no error ", func() {
			err := stateValidator.Validate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when state file cannot be found", func() {
		It("returns an error", func() {
			err := stateValidator.Validate()
			expectedError := fmt.Errorf("bbl-state.json not found in %q, ensure you're running this command in the proper state directory or create a new environment with bbl up", tempDirectory)
			Expect(err).To(MatchError(expectedError))
		})
	})

	Context("failure cases", func() {
		Context("when permission denied", func() {
			BeforeEach(func() {
				err := os.Chmod(tempDirectory, os.FileMode(0))
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				err := stateValidator.Validate()
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})
})
