package application_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/application"

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

	It("returns no error when state file exists", func() {
		err := ioutil.WriteFile(filepath.Join(tempDirectory, "bbl-state.json"), []byte(""), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		err = stateValidator.Validate()
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns an error when state file cannot be found", func() {
		err := stateValidator.Validate()
		expectedError := fmt.Errorf("bbl-state.json not found in %q, ensure you're running this command in the proper state directory or create a new environment with bbl up", tempDirectory)
		Expect(err).To(MatchError(expectedError))
	})

	Context("failure cases", func() {
		It("returns an error when permission denied", func() {
			err := os.Chmod(tempDirectory, os.FileMode(0))
			Expect(err).NotTo(HaveOccurred())

			err = stateValidator.Validate()
			Expect(err).To(MatchError(ContainSubstring("permission denied")))
		})
	})
})
