package bosh_test

import (
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
			state         storage.State
			expectedState storage.State
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
			state = storage.State{
				BOSH: storage.BOSH{
					Variables: beforeDeletionVars,
				},
				Jumpbox: storage.Jumpbox{
					Variables: beforeDeletionVars,
				},
			}
			expectedState = storage.State{
				BOSH: storage.BOSH{
					Variables: afterDeletionVars,
				},
				Jumpbox: storage.Jumpbox{
					Variables: afterDeletionVars,
				},
			}
		})

		It("deletes the jumpbox ssh key from the state and returns the new state", func() {
			newState, err := sshKeyDeleter.Delete(state)
			Expect(err).NotTo(HaveOccurred())
			Expect(newState).To(Equal(expectedState))
		})

		Context("when the jumpbox-variables.yml exists on disk", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(varsDir, "jumpbox-variables.yml"), []byte(beforeDeletionVars), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})
			It("deletes the jumpbox key from the file on disk", func() {
				_, err := sshKeyDeleter.Delete(storage.State{})
				Expect(err).NotTo(HaveOccurred())

				contents, err := ioutil.ReadFile(filepath.Join(varsDir, "jumpbox-variables.yml"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).To(Equal(afterDeletionVars))
			})
		})

		Context("when the jumpbox-variables.yml does not exists on disk", func() {
			It("does not write the file to disk", func() {
				_, err := sshKeyDeleter.Delete(state)
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Stat(filepath.Join(varsDir, "jumpbox-variables.yml"))
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the BOSH variables is invalid YAML", func() {
			It("returns an error", func() {
				state.BOSH.Variables = "invalid yaml"
				_, err := sshKeyDeleter.Delete(state)
				Expect(err).To(MatchError(ContainSubstring("BOSH variables: yaml: unmarshal errors:")))
			})
		})

		Context("when the Jumpbox variables is invalid YAML", func() {
			It("returns an error", func() {
				state.Jumpbox.Variables = "invalid yaml"
				_, err := sshKeyDeleter.Delete(state)
				Expect(err).To(MatchError(ContainSubstring("Jumpbox variables: yaml: unmarshal errors:")))
			})
		})
	})
})
