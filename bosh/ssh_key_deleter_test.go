package bosh_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSHKeyDeleter", func() {
	Describe("Delete", func() {
		var (
			sshKeyDeleter bosh.SSHKeyDeleter
			varsDir       string
			stateStore    *fakes.StateStore

			beforeDeletionVars string
			afterDeletionVars  string
		)

		BeforeEach(func() {
			var err error
			varsDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			stateStore = &fakes.StateStore{}
			stateStore.GetVarsDirCall.Returns.Directory = varsDir

			beforeDeletionVars = "foo: bar\njumpbox_ssh:\n  private_key: some-private-key"
			afterDeletionVars = "foo: bar\n"

			sshKeyDeleter = bosh.NewSSHKeyDeleter(stateStore)
		})

		Context("when the jumpbox-variables.yml exists on disk", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(varsDir, "jumpbox-variables.yml"), []byte(beforeDeletionVars), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())
			})
			It("deletes the jumpbox key from the file on disk", func() {
				err := sshKeyDeleter.Delete()
				Expect(err).NotTo(HaveOccurred())

				contents, err := ioutil.ReadFile(filepath.Join(varsDir, "jumpbox-variables.yml"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal(afterDeletionVars))
			})
		})

		Context("when the jumpbox-variables.yml doesn't contain ssh key", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(varsDir, "jumpbox-variables.yml"), []byte(afterDeletionVars), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())
			})
			It("does nothing", func() {
				err := sshKeyDeleter.Delete()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the jumpbox-variables.yml does not exist on disk", func() {
			It("does not write the file to disk", func() {
				err := sshKeyDeleter.Delete()
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Stat(filepath.Join(varsDir, "jumpbox-variables.yml"))
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the Jumpbox variables is invalid YAML", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(varsDir, "jumpbox-variables.yml"), []byte("invalid yaml"), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				err := sshKeyDeleter.Delete()
				Expect(err).To(MatchError(ContainSubstring("Jumpbox variables: yaml: unmarshal errors:")))
			})
		})

		Context("when the vars dir can't be accessed", func() {
			BeforeEach(func() {
				stateStore.GetVarsDirCall.Returns.Error = errors.New("potato")
			})
			It("returns an error", func() {
				err := sshKeyDeleter.Delete()
				Expect(err).To(MatchError("Get vars dir: potato"))
			})
		})
	})
})
