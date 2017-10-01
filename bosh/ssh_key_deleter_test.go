package bosh_test

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSHKeyDeleter", func() {
	Describe("Delete", func() {
		var (
			sshKeyDeleter bosh.SSHKeyDeleter
			state         storage.State
			expectedState storage.State
		)

		BeforeEach(func() {
			sshKeyDeleter = bosh.NewSSHKeyDeleter()
			state = storage.State{
				BOSH: storage.BOSH{
					Variables: "foo: bar\njumpbox_ssh:\n  private_key: some-private-key",
				},
				Jumpbox: storage.Jumpbox{
					Variables: "foo: bar\njumpbox_ssh:\n  private_key: some-private-key",
				},
			}
			expectedState = storage.State{
				BOSH: storage.BOSH{
					Variables: "foo: bar\n",
				},
				Jumpbox: storage.Jumpbox{
					Variables: "foo: bar\n",
				},
			}
		})

		It("deletes the jumpbox ssh key from the state and returns the new state", func() {
			newState, err := sshKeyDeleter.Delete(state)
			Expect(err).NotTo(HaveOccurred())
			Expect(newState).To(Equal(expectedState))
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
