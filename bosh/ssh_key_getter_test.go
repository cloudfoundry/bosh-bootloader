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
			state = storage.State{}
		})

		Context("when there is a jumpbox", func() {
			BeforeEach(func() {
				state = storage.State{
					Jumpbox: storage.Jumpbox{
						Enabled:   true,
						Variables: "jumpbox_ssh:\n  private_key: some-private-key",
					},
				}
			})

			It("returns the jumpbox ssh key from the state", func() {
				privateKey, err := sshKeyGetter.Get(state)
				Expect(err).NotTo(HaveOccurred())
				Expect(privateKey).To(Equal("some-private-key"))
			})
		})

		Context("when there is no jumpbox", func() {
			Context("when there is a top-level keypair", func() {
				BeforeEach(func() {
					state = storage.State{
						BOSH: storage.BOSH{
							Variables: "some-var: wut",
						},
						KeyPair: storage.KeyPair{
							PrivateKey: "some-private-key",
							PublicKey:  "some-public-key",
						},
					}
				})

				It("returns the ssh key from the state", func() {
					privateKey, err := sshKeyGetter.Get(state)
					Expect(err).NotTo(HaveOccurred())
					Expect(privateKey).To(Equal("some-private-key"))
				})
			})

			Context("when keypair is in bosh variables", func() {
				BeforeEach(func() {
					state = storage.State{
						BOSH: storage.BOSH{
							Variables: "jumpbox_ssh:\n  private_key: some-private-key",
						},
						KeyPair: storage.KeyPair{
							PrivateKey: "some-old-private-key",
							PublicKey:  "some-public-key",
						},
					}
				})

				It("returns the jumpbox ssh key from the state", func() {
					privateKey, err := sshKeyGetter.Get(state)
					Expect(err).NotTo(HaveOccurred())
					Expect(privateKey).To(Equal("some-private-key"))
				})
			})
		})

		Context("failure cases", func() {
			It("returns an error when the BOSH variables yaml cannot be unmarshaled", func() {
				state.BOSH.Variables = "invalid yaml"
				_, err := sshKeyGetter.Get(state)
				Expect(err).To(MatchError(ContainSubstring("line 1: cannot unmarshal !!str `invalid...`")))
			})
		})
	})
})
