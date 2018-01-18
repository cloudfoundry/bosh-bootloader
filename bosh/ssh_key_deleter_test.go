package bosh_test

import (
	"errors"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSHKeyDeleter", func() {
	Describe("Delete", func() {
		var (
			sshKeyDeleter bosh.SSHKeyDeleter
			stateStore    *fakes.StateStore
			fileIO        *fakes.FileIO

			beforeDeletionVars string
			afterDeletionVars  string
		)

		BeforeEach(func() {
			stateStore = &fakes.StateStore{}
			stateStore.GetVarsDirCall.Returns.Directory = "some-vars-dir"

			fileIO = &fakes.FileIO{}

			beforeDeletionVars = "foo: bar\njumpbox_ssh:\n  private_key: some-private-key"
			afterDeletionVars = "foo: bar\n"

			sshKeyDeleter = bosh.NewSSHKeyDeleter(stateStore, fileIO)
		})

		Context("when the jumpbox-vars-store.yml exists on disk", func() {
			BeforeEach(func() {
				fileIO.ReadFileCall.Returns.Contents = []byte(beforeDeletionVars)
			})
			It("deletes the jumpbox key from the file on disk", func() {
				err := sshKeyDeleter.Delete()
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.ReadFileCall.Receives.Filename).To(Equal(filepath.Join("some-vars-dir", "jumpbox-vars-store.yml")))

				Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(filepath.Join("some-vars-dir", "jumpbox-vars-store.yml")))
				Expect(string(fileIO.WriteFileCall.Receives[0].Contents)).To(Equal(afterDeletionVars))
			})
		})

		Context("when the jumpbox-vars-store.yml doesn't contain ssh key", func() {
			BeforeEach(func() {
				fileIO.ReadFileCall.Returns.Contents = []byte(afterDeletionVars)
			})
			It("does nothing", func() {
				err := sshKeyDeleter.Delete()
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.WriteFileCall.CallCount).To(Equal(0))
			})
		})

		Context("when the jumpbox-vars-store.yml does not exist on disk", func() {
			BeforeEach(func() {
				fileIO.ReadFileCall.Returns.Error = errors.New("some read error")
			})
			It("does not write the file to disk", func() {
				err := sshKeyDeleter.Delete()
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.WriteFileCall.CallCount).To(Equal(0))
			})
		})

		Context("when the Jumpbox variables is invalid YAML", func() {
			BeforeEach(func() {
				fileIO.ReadFileCall.Returns.Contents = []byte("invalid yaml")
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
