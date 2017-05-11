package aws_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/keypair"
	"github.com/cloudfoundry/bosh-bootloader/keypair/aws"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {
	Describe("Rotate", func() {
		var (
			keyPairSynchronizer *fakes.KeyPairSynchronizer
			keyPairManager      aws.Manager
		)

		BeforeEach(func() {
			keyPairSynchronizer = &fakes.KeyPairSynchronizer{}

			keyPairManager = aws.NewManager(keyPairSynchronizer)
		})

		It("returns a helpful error message", func() {
			_, err := keyPairManager.Rotate(storage.State{})
			Expect(err).To(MatchError("rotating aws keys is not yet implemented"))
		})
	})

	Describe("Sync", func() {
		var (
			keyPairSynchronizer *fakes.KeyPairSynchronizer

			keyPairManager aws.Manager

			incomingState storage.State
		)

		BeforeEach(func() {
			keyPairSynchronizer = &fakes.KeyPairSynchronizer{}
			incomingState = storage.State{
				EnvID: "some-env-id",
			}

			keyPairManager = aws.NewManager(keyPairSynchronizer)
		})

		It("generates a keypair name if one doesn't exist", func() {
			updatedState, err := keyPairManager.Sync(incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedState).To(Equal(storage.State{
				EnvID: "some-env-id",
				KeyPair: storage.KeyPair{
					Name: "keypair-some-env-id",
				},
			}))
		})

		It("syncs the keypair", func() {
			incomingState.KeyPair = storage.KeyPair{
				Name:       "some-keypair-name",
				PrivateKey: "some-private-key",
				PublicKey:  "some-public-key",
			}
			_, err := keyPairManager.Sync(incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(keyPairSynchronizer.SyncCall.CallCount).To(Equal(1))
			Expect(keyPairSynchronizer.SyncCall.Receives.KeyPair).To(Equal(ec2.KeyPair{
				Name:       "some-keypair-name",
				PrivateKey: "some-private-key",
				PublicKey:  "some-public-key",
			}))
		})

		It("saves the keypair to the state", func() {
			keyPairSynchronizer.SyncCall.Returns.KeyPair = ec2.KeyPair{
				Name:       "some-keypair-name",
				PrivateKey: "some-private-key",
				PublicKey:  "some-public-key",
			}
			updatedState, err := keyPairManager.Sync(incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedState).To(Equal(storage.State{
				EnvID: "some-env-id",
				KeyPair: storage.KeyPair{
					Name:       "keypair-some-env-id",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
			}))
		})

		Context("when a keypair name already exists", func() {
			It("reuses that keypair name", func() {
				incomingState.KeyPair.Name = "some-other-keypair-name"
				updatedState, err := keyPairManager.Sync(incomingState)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedState).To(Equal(storage.State{
					EnvID: "some-env-id",
					KeyPair: storage.KeyPair{
						Name: "some-other-keypair-name",
					},
				}))
			})
		})

		Context("failure cases", func() {
			Context("when the state doesn't have an env id", func() {
				It("returns an error", func() {
					_, err := keyPairManager.Sync(storage.State{})
					Expect(err).To(MatchError("env id must be set to generate a keypair"))
				})
			})

			Context("when the keypair synchronizer fails", func() {
				It("returns a manager error", func() {
					keyPairSynchronizer.SyncCall.Returns.Error = errors.New("failed to sync")
					incomingState.KeyPair.Name = "some-keypair-name"
					_, err := keyPairManager.Sync(incomingState)
					Expect(err).To(MatchError(keypair.NewManagerError(incomingState, errors.New("failed to sync"))))
				})
			})
		})
	})
})
