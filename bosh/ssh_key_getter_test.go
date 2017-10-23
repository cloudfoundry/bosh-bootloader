package bosh_test

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SSHKeyGetter", func() {
	Describe("Get", func() {
		var (
			sshKeyGetter bosh.SSHKeyGetter
			variables    string
		)

		BeforeEach(func() {
			sshKeyGetter = bosh.NewSSHKeyGetter()
			variables = "jumpbox_ssh:\n  private_key: some-private-key"
		})

		It("returns the jumpbox ssh key from the state", func() {
			privateKey, err := sshKeyGetter.Get(variables)
			Expect(err).NotTo(HaveOccurred())
			Expect(privateKey).To(Equal("some-private-key"))
		})

		Context("failure cases", func() {
			Context("when the Jumpbox variables yaml cannot be unmarshaled", func() {
				It("returns an error", func() {
					_, err := sshKeyGetter.Get("invalid yaml")
					Expect(err).To(MatchError(ContainSubstring("line 1: cannot unmarshal !!str `invalid...`")))
				})
			})
		})
	})
})
