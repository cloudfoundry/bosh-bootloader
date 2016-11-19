package terraform_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("Applier", func() {
	var (
		cmd     *fakes.TerraformCmd
		applier terraform.Applier
		tempDir string
	)

	BeforeEach(func() {
		cmd = &fakes.TerraformCmd{}

		applier = terraform.NewApplier(cmd)

		terraform.SetTempDir(func(dir, prefix string) (string, error) {
			var err error
			tempDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			return tempDir, nil
		})

		terraform.SetReadFile(func(string) ([]byte, error) {
			return []byte(""), nil
		})
	})

	AfterEach(func() {
		terraform.ResetTempDir()
		terraform.ResetReadFile()
	})

	It("terraform command is called with apply", func() {
		_, err := applier.Apply("some/credential/file", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template")
		Expect(err).NotTo(HaveOccurred())

		Expect(cmd.RunCall.Receives.WorkingDirectory).To(Equal(tempDir))
		Expect(cmd.RunCall.Receives.Args[0]).To(Equal("apply"))
	})

	It("saves the terraform template to disk", func() {
		_, err := applier.Apply("some/credential/file", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template")
		Expect(err).NotTo(HaveOccurred())

		Expect(len(cmd.RunCall.Receives.Args)).NotTo(Equal(0))
		path := cmd.RunCall.Receives.Args[len(cmd.RunCall.Receives.Args)-1]

		fileContents, err := ioutil.ReadFile(filepath.Join(path, "template.tf"))
		Expect(err).NotTo(HaveOccurred())

		Expect(string(fileContents)).To(Equal("some-template"))
	})

	It("passes vars to apply", func() {
		_, err := applier.Apply("some/credential/file", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template")
		Expect(err).NotTo(HaveOccurred())

		Expect(cmd.RunCall.Receives.Args).To(ContainSequence([]string{"-var", "credentials=some/credential/file"}))
		Expect(cmd.RunCall.Receives.Args).To(ContainSequence([]string{"-var", "env_id=some-env-id"}))
		Expect(cmd.RunCall.Receives.Args).To(ContainSequence([]string{"-var", "project_id=some-project-id"}))
		Expect(cmd.RunCall.Receives.Args).To(ContainSequence([]string{"-var", "zone=some-zone"}))
		Expect(cmd.RunCall.Receives.Args).To(ContainSequence([]string{"-var", "region=some-region"}))
	})

	Context("when the terraform command saves a terraform.tfstate", func() {
		var (
			actualFilename string
		)

		BeforeEach(func() {
			terraform.SetReadFile(func(filename string) ([]byte, error) {
				actualFilename = filename
				return []byte("some-terraform-state"), nil
			})
		})

		AfterEach(func() {
			terraform.ResetReadFile()
		})

		It("returns the terraform state", func() {
			terraformState, err := applier.Apply("some/credential/file", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template")
			Expect(err).NotTo(HaveOccurred())

			Expect(actualFilename).To(ContainSubstring("terraform.tfstate"))
			Expect(terraformState).To(Equal("some-terraform-state"))
		})
	})

	Context("failure case", func() {
		Context("when it fails to create a temp dir", func() {
			BeforeEach(func() {
				terraform.SetTempDir(func(dir, prefix string) (string, error) {
					return "", errors.New("failed to make temp dir")
				})
			})

			AfterEach(func() {
				terraform.ResetTempDir()
			})

			It("returns an error", func() {
				_, err := applier.Apply("some/credential/file", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template")
				Expect(err).To(MatchError("failed to make temp dir"))
			})
		})

		Context("when it fails to write a file", func() {
			BeforeEach(func() {
				terraform.SetWriteFile(func(file string, data []byte, perm os.FileMode) error {
					return errors.New("failed to write a file")
				})
			})

			AfterEach(func() {
				terraform.ResetWriteFile()
			})

			It("returns an error", func() {
				_, err := applier.Apply("some/credential/file", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template")
				Expect(err).To(MatchError("failed to write a file"))
			})
		})

		Context("when it fails to call terraform command run", func() {
			It("returns an error", func() {
				cmd.RunCall.Returns.Error = errors.New("failed to run terraform command")
				_, err := applier.Apply("some/credential/file", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template")
				Expect(err).To(MatchError("failed to run terraform command"))
			})
		})

		Context("when it fails to read a file", func() {
			BeforeEach(func() {
				terraform.SetReadFile(func(filename string) ([]byte, error) {
					return []byte{}, errors.New("failed to read a file")
				})
			})

			AfterEach(func() {
				terraform.ResetReadFile()
			})

			It("returns an error", func() {
				_, err := applier.Apply("some/credential/file", "some-env-id", "some-project-id", "some-zone", "some-region", "some-template")
				Expect(err).To(MatchError("failed to read a file"))
			})
		})
	})
})
