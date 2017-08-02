package gcp_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/keypair/gcp"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {
	Describe("Sync", func() {
		var (
			keyPairUpdater *fakes.GCPKeyPairUpdater
			keyPairDeleter *fakes.GCPKeyPairDeleter
			keyPairManager gcp.Manager
		)

		BeforeEach(func() {
			keyPairUpdater = &fakes.GCPKeyPairUpdater{}
			keyPairUpdater.UpdateCall.Returns.KeyPair = storage.KeyPair{
				PrivateKey: "some-private-key",
				PublicKey:  "some-public-key",
			}
			keyPairDeleter = &fakes.GCPKeyPairDeleter{}

			keyPairManager = gcp.NewManager(keyPairUpdater, keyPairDeleter)
		})

		Context("when keypair is empty", func() {
			It("calls keypair updater and saves the keypair to the state", func() {
				state, err := keyPairManager.Sync(storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(keyPairUpdater.UpdateCall.CallCount).To(Equal(1))
				Expect(state).To(Equal(storage.State{
					KeyPair: storage.KeyPair{
						PrivateKey: "some-private-key",
						PublicKey:  "some-public-key",
					},
				}))
			})

			Context("failure cases", func() {
				It("returns an error when key pair updater fails", func() {
					keyPairUpdater.UpdateCall.Returns.Error = errors.New("failed to update")
					_, err := keyPairManager.Sync(storage.State{})
					Expect(err).To(MatchError("failed to update"))
				})
			})
		})

		Context("when keypair exists", func() {
			It("no-ops and returns provided state", func() {
				state, err := keyPairManager.Sync(storage.State{
					KeyPair: storage.KeyPair{
						PrivateKey: "some-existing-private-key",
						PublicKey:  "some-existing-public-key",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(keyPairUpdater.UpdateCall.CallCount).To(Equal(0))
				Expect(state).To(Equal(storage.State{
					KeyPair: storage.KeyPair{
						PrivateKey: "some-existing-private-key",
						PublicKey:  "some-existing-public-key",
					},
				}))
			})
		})
	})

	Describe("Rotate", func() {
		var (
			keyPairUpdater *fakes.GCPKeyPairUpdater
			keyPairDeleter *fakes.GCPKeyPairDeleter
			keyPairManager gcp.Manager
		)

		BeforeEach(func() {
			keyPairUpdater = &fakes.GCPKeyPairUpdater{}
			keyPairUpdater.UpdateCall.Returns.KeyPair = storage.KeyPair{
				PrivateKey: "some-new-private-key",
				PublicKey:  "some-new-public-key",
			}
			keyPairDeleter = &fakes.GCPKeyPairDeleter{}
			keyPairManager = gcp.NewManager(keyPairUpdater, keyPairDeleter)
		})

		Context("when keypair is empty", func() {
			It("returns a helpful error message", func() {
				_, err := keyPairManager.Rotate(storage.State{})
				Expect(err).To(MatchError("no key found to rotate"))
			})
		})

		Context("when keypair exists", func() {
			It("deletes the old keypair and creates the new keypair", func() {
				state, err := keyPairManager.Rotate(storage.State{
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
					KeyPair: storage.KeyPair{
						PrivateKey: "some-existing-private-key",
						PublicKey:  "some-existing-public-key",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(keyPairDeleter.DeleteCall.CallCount).To(Equal(1))
				Expect(keyPairDeleter.DeleteCall.Receives.PublicKey).To(Equal("some-existing-public-key"))
				Expect(keyPairUpdater.UpdateCall.CallCount).To(Equal(1))

				Expect(state).To(Equal(storage.State{
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
					},
					KeyPair: storage.KeyPair{
						PrivateKey: "some-new-private-key",
						PublicKey:  "some-new-public-key",
					},
				}))
			})
		})

		Context("failure cases", func() {
			It("returns an error when key pair deleter delete fails", func() {
				keyPairDeleter.DeleteCall.Returns.Error = errors.New("failed to delete")
				_, err := keyPairManager.Rotate(storage.State{
					KeyPair: storage.KeyPair{
						PrivateKey: "some-existing-private-key",
						PublicKey:  "some-existing-public-key",
					},
				})
				Expect(err).To(MatchError("failed to delete"))
			})

			It("returns an error when the key pair updater update fails", func() {
				keyPairUpdater.UpdateCall.Returns.Error = errors.New("failed to update")
				_, err := keyPairManager.Rotate(storage.State{
					KeyPair: storage.KeyPair{
						PrivateKey: "some-existing-private-key",
						PublicKey:  "some-existing-public-key",
					},
				})
				Expect(err).To(MatchError("failed to update"))
			})
		})
	})
})
