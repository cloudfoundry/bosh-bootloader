package bosh_test

import (
	"errors"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSHKeyGetter", func() {
	Describe("Get", func() {
		var (
			sshKeyGetter bosh.SSHKeyGetter
			stateStore   *fakes.StateStore
			fileIO       *fakes.FileIO
			variables    string
		)

		BeforeEach(func() {
			stateStore = &fakes.StateStore{}
			stateStore.GetVarsDirCall.Returns.Directory = "some-fake-vars-dir"

			fileIO = &fakes.FileIO{}
			variables = "jumpbox_ssh:\n  private_key: some-private-key"
			fileIO.ReadFileCall.Returns.Contents = []byte(variables)

			sshKeyGetter = bosh.NewSSHKeyGetter(stateStore, fileIO)
		})

		It("returns the jumpbox ssh key from the state", func() {
			privateKey, err := sshKeyGetter.Get("some-deployment")
			Expect(err).NotTo(HaveOccurred())
			Expect(privateKey).To(Equal("some-private-key"))
			Expect(fileIO.ReadFileCall.Receives.Filename).To(Equal(filepath.Join("some-fake-vars-dir", "some-deployment-vars-store.yml")))
		})

		Context("failure cases", func() {
			Context("when the Jumpbox variables yaml cannot be unmarshaled", func() {
				BeforeEach(func() {
					fileIO.ReadFileCall.Returns.Contents = []byte("invalid yaml")
				})

				It("returns an error", func() {
					_, err := sshKeyGetter.Get("invalid-deployment")
					Expect(err).To(MatchError(ContainSubstring("line 1: cannot unmarshal !!str `invalid...`")))
				})
			})

			Context("when the state store fails to get the vars dir", func() {
				BeforeEach(func() {
					stateStore.GetVarsDirCall.Returns.Error = errors.New("tangelo")
				})

				It("returns an error", func() {
					_, err := sshKeyGetter.Get("some-deployment")
					Expect(err).To(MatchError("Get vars directory: tangelo"))
				})
			})

			Context("when the deployment vars file can't be read", func() {
				BeforeEach(func() {
					fileIO.ReadFileCall.Returns.Error = errors.New("orange")
				})

				It("returns an error", func() {
					_, err := sshKeyGetter.Get("nonexistent")
					Expect(err).To(MatchError(ContainSubstring("Read nonexistent vars file: ")))
					Expect(err).To(MatchError(ContainSubstring("orange")))
				})
			})
		})
	})
})
