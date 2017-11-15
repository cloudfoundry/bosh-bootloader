package bosh_test

import (
	"errors"
	"io/ioutil"
	"os"
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
			variables    string
			stateStore   *fakes.StateStore
			varsDir      string
		)

		BeforeEach(func() {
			var err error
			varsDir, err = ioutil.TempDir("", "")
			stateStore = &fakes.StateStore{}
			stateStore.GetVarsDirCall.Returns.Directory = varsDir
			sshKeyGetter = bosh.NewSSHKeyGetter(stateStore)
			variables = "jumpbox_ssh:\n  private_key: some-private-key"
			err = ioutil.WriteFile(filepath.Join(varsDir, "some-deployment-variables.yml"), []byte(variables), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the jumpbox ssh key from the state", func() {
			privateKey, err := sshKeyGetter.Get("some-deployment")
			Expect(err).NotTo(HaveOccurred())
			Expect(privateKey).To(Equal("some-private-key"))
		})

		Context("failure cases", func() {
			Context("when the Jumpbox variables yaml cannot be unmarshaled", func() {
				BeforeEach(func() {
					err := ioutil.WriteFile(filepath.Join(varsDir, "invalid-deployment-variables.yml"), []byte("invalid yaml"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())
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
				It("returns an error", func() {
					_, err := sshKeyGetter.Get("nonexistent")
					Expect(err).To(MatchError(ContainSubstring("Read nonexistent vars file: ")))
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})
		})
	})
})
