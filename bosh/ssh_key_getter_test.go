package bosh_test

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSHKeyGetter", func() {
	Describe("Get", func() {
		var (
			sshKeyGetter bosh.SSHKeyGetter
			state        storage.State
		)

		BeforeEach(func() {
			sshKeyGetter = bosh.NewSSHKeyGetter()
			state = storage.State{
				Jumpbox: storage.Jumpbox{
					Variables: "jumpbox_ssh:\n  private_key: some-private-key",
				},
			}
		})

		It("returns the jumpbox ssh key from the state", func() {
			privateKey, err := sshKeyGetter.Get(state)
			Expect(err).NotTo(HaveOccurred())
			Expect(privateKey).To(Equal("some-private-key"))
		})

		Context("failure cases", func() {
			It("returns an error when the Jumpbox variables yaml cannot be unmarshaled", func() {
				state.Jumpbox.Variables = "invalid yaml"
				_, err := sshKeyGetter.Get(state)
				Expect(err).To(MatchError(ContainSubstring("line 1: cannot unmarshal !!str `invalid...`")))
			})
		})
	})
})
