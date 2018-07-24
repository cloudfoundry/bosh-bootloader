package application_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

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
			Expect(stateValidator.Validate()).To(Succeed())
		})
	})

	Context("when state file cannot be found", func() {
		It("returns an error", func() {
			Expect(stateValidator.Validate()).To(MatchError(fmt.Sprintf("bbl-state.json not found in %q, ensure you're running this command in the proper state directory or create a new environment with bbl up", tempDirectory)))
		})
	})

	Context("failure cases", func() {
		Context("when permission denied", func() {
			It("returns an error", func() {
				if runtime.GOOS == "windows" {
					Skip("Chmod is not supported on Windows")
				}
				Expect(os.Chmod(tempDirectory, os.FileMode(0))).To(Succeed())

				Expect(stateValidator.Validate()).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})
})
