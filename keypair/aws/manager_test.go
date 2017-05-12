package aws_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/keypair"
	keypairaws "github.com/cloudfoundry/bosh-bootloader/keypair/aws"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {
	Describe("Rotate", func() {
		var (
			awsKeyPairSynchronizer *fakes.AWSKeyPairSynchronizer
			awsKeyPairDeleter      *fakes.AWSKeyPairDeleter
			awsClientProvider      *fakes.AWSClientProvider
			keyPairManager         keypairaws.Manager
		)

		BeforeEach(func() {
			awsKeyPairSynchronizer = &fakes.AWSKeyPairSynchronizer{}
			awsKeyPairDeleter = &fakes.AWSKeyPairDeleter{}
			awsClientProvider = &fakes.AWSClientProvider{}

			awsKeyPairSynchronizer.SyncCall.Returns.KeyPair = ec2.KeyPair{
				Name:       "some-keypair-name",
				PrivateKey: "some-new-private-key",
				PublicKey:  "some-new-public-key",
			}

			keyPairManager = keypairaws.NewManager(awsKeyPairSynchronizer, awsKeyPairDeleter, awsClientProvider)
		})

		Context("when the keypair is empty", func() {
			It("returns a helpful error message", func() {
				_, err := keyPairManager.Rotate(storage.State{})
				Expect(err).To(MatchError("no key found to rotate"))
			})
		})

		Context("when keypair exists", func() {
			It("deletes the old keypair and creates a new keypair", func() {
				state, err := keyPairManager.Rotate(storage.State{
					AWS: storage.AWS{
						SecretAccessKey: "some-secret-access-key",
						AccessKeyID:     "some-access-key-id",
						Region:          "some-region",
					},
					KeyPair: storage.KeyPair{
						Name:       "some-keypair-name",
						PrivateKey: "some-existing-private-key",
						PublicKey:  "some-existing-public-key",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(awsClientProvider.SetConfigCall.CallCount).To(Equal(1))
				Expect(awsClientProvider.SetConfigCall.Receives.Config).To(Equal(aws.Config{
					AccessKeyID:     "some-access-key-id",
					SecretAccessKey: "some-secret-access-key",
					Region:          "some-region",
				}))
				Expect(awsKeyPairDeleter.DeleteCall.CallCount).To(Equal(1))
				Expect(awsKeyPairDeleter.DeleteCall.Receives.Name).To(Equal("some-keypair-name"))
				Expect(awsKeyPairSynchronizer.SyncCall.CallCount).To(Equal(1))
				Expect(awsKeyPairSynchronizer.SyncCall.Receives.KeyPair).To(Equal(ec2.KeyPair{
					Name: "some-keypair-name",
				}))

				Expect(state).To(Equal(storage.State{
					AWS: storage.AWS{
						SecretAccessKey: "some-secret-access-key",
						AccessKeyID:     "some-access-key-id",
						Region:          "some-region",
					},
					KeyPair: storage.KeyPair{
						Name:       "some-keypair-name",
						PrivateKey: "some-new-private-key",
						PublicKey:  "some-new-public-key",
					},
				}))
			})

		})

		Context("failure cases", func() {
			It("returns an error when key pair deleter delete fails", func() {
				awsKeyPairDeleter.DeleteCall.Returns.Error = errors.New("key pair deleter delete failed")
				_, err := keyPairManager.Rotate(storage.State{
					KeyPair: storage.KeyPair{
						Name:       "some-keypair-name",
						PrivateKey: "some-existing-private-key",
						PublicKey:  "some-existing-public-key",
					},
				})
				Expect(err).To(MatchError("key pair deleter delete failed"))
			})

			It("returns an error when key pair synchornizer sync fails", func() {
				awsKeyPairSynchronizer.SyncCall.Returns.Error = errors.New("key pair synchronizer sync failed")
				_, err := keyPairManager.Rotate(storage.State{
					KeyPair: storage.KeyPair{
						Name:       "some-keypair-name",
						PrivateKey: "some-existing-private-key",
						PublicKey:  "some-existing-public-key",
					},
				})
				Expect(err).To(MatchError("key pair synchronizer sync failed"))
			})
		})
	})

	Describe("Sync", func() {
		var (
			awsKeyPairSynchronizer *fakes.AWSKeyPairSynchronizer
			awsKeyPairDeleter      *fakes.AWSKeyPairDeleter
			awsClientProvider      *fakes.AWSClientProvider

			keyPairManager keypairaws.Manager

			incomingState storage.State
		)

		BeforeEach(func() {
			awsKeyPairSynchronizer = &fakes.AWSKeyPairSynchronizer{}
			awsKeyPairDeleter = &fakes.AWSKeyPairDeleter{}
			awsClientProvider = &fakes.AWSClientProvider{}
			incomingState = storage.State{
				EnvID: "some-env-id",
			}

			keyPairManager = keypairaws.NewManager(awsKeyPairSynchronizer, awsKeyPairDeleter, awsClientProvider)
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
			Expect(awsKeyPairSynchronizer.SyncCall.CallCount).To(Equal(1))
			Expect(awsKeyPairSynchronizer.SyncCall.Receives.KeyPair).To(Equal(ec2.KeyPair{
				Name:       "some-keypair-name",
				PrivateKey: "some-private-key",
				PublicKey:  "some-public-key",
			}))
		})

		It("saves the keypair to the state", func() {
			awsKeyPairSynchronizer.SyncCall.Returns.KeyPair = ec2.KeyPair{
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
					awsKeyPairSynchronizer.SyncCall.Returns.Error = errors.New("failed to sync")
					incomingState.KeyPair.Name = "some-keypair-name"
					_, err := keyPairManager.Sync(incomingState)
					Expect(err).To(MatchError(keypair.NewManagerError(incomingState, errors.New("failed to sync"))))
				})
			})
		})
	})
})
