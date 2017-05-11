package keypair_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/keypair"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {
	Describe("Sync", func() {
		var (
			awsManager *fakes.KeyPairManager
			gcpManager *fakes.KeyPairManager

			keyPairManager keypair.Manager
		)

		BeforeEach(func() {
			awsManager = &fakes.KeyPairManager{}
			gcpManager = &fakes.KeyPairManager{}

			awsManager.SyncCall.Returns.State = storage.State{
				KeyPair: storage.KeyPair{
					Name: "some-aws-keypair",
				},
			}

			gcpManager.SyncCall.Returns.State = storage.State{
				KeyPair: storage.KeyPair{
					Name: "some-gcp-keypair",
				},
			}

			keyPairManager = keypair.NewManager(awsManager, gcpManager)
		})

		Context("when iaas is aws", func() {
			It("calls the aws manager sync and returns state", func() {
				state, err := keyPairManager.Sync(storage.State{
					IAAS: "aws",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(awsManager.SyncCall.CallCount).To(Equal(1))
				Expect(awsManager.SyncCall.Receives.State).To(Equal(storage.State{
					IAAS: "aws",
				}))

				Expect(state).To(Equal(storage.State{
					IAAS: "aws",
					KeyPair: storage.KeyPair{
						Name: "some-aws-keypair",
					},
				}))
			})

			Context("failure cases", func() {
				It("returns an error when sync fails", func() {
					awsManager.SyncCall.Returns.Error = errors.New("call to sync failed")
					_, err := keyPairManager.Sync(storage.State{
						IAAS: "aws",
					})
					Expect(err).To(MatchError("call to sync failed"))
				})
			})
		})

		Context("when iaas is gcp", func() {
			It("calls the gcp manager sync and returns state", func() {
				state, err := keyPairManager.Sync(storage.State{
					IAAS: "gcp",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(gcpManager.SyncCall.CallCount).To(Equal(1))
				Expect(gcpManager.SyncCall.Receives.State).To(Equal(storage.State{
					IAAS: "gcp",
				}))

				Expect(state).To(Equal(storage.State{
					IAAS: "gcp",
					KeyPair: storage.KeyPair{
						Name: "some-gcp-keypair",
					},
				}))
			})

			Context("failure cases", func() {
				It("returns an error when sync fails", func() {
					gcpManager.SyncCall.Returns.Error = errors.New("call to sync failed")
					_, err := keyPairManager.Sync(storage.State{
						IAAS: "gcp",
					})
					Expect(err).To(MatchError("call to sync failed"))
				})
			})
		})

		Context("failure cases", func() {
			Context("when iaas is invalid", func() {
				It("returns an error", func() {
					_, err := keyPairManager.Sync(storage.State{
						IAAS: "invalid-iaas",
					})
					Expect(err).To(MatchError("invalid iaas was provided: invalid-iaas"))
				})
			})
		})
	})

	Describe("Rotate", func() {
		var (
			awsManager *fakes.KeyPairManager
			gcpManager *fakes.KeyPairManager

			keyPairManager keypair.Manager
		)

		BeforeEach(func() {
			awsManager = &fakes.KeyPairManager{}
			gcpManager = &fakes.KeyPairManager{}

			awsManager.RotateCall.Returns.State = storage.State{
				KeyPair: storage.KeyPair{
					Name: "some-new-aws-keypair",
				},
			}

			gcpManager.RotateCall.Returns.State = storage.State{
				KeyPair: storage.KeyPair{
					Name: "some-new-gcp-keypair",
				},
			}

			keyPairManager = keypair.NewManager(awsManager, gcpManager)
		})

		Context("when iaas is aws", func() {
			It("calls the aws manager rotate and returns state", func() {
				state, err := keyPairManager.Rotate(storage.State{
					IAAS: "aws",
					KeyPair: storage.KeyPair{
						Name: "some-aws-keypair",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(awsManager.RotateCall.CallCount).To(Equal(1))
				Expect(awsManager.RotateCall.Receives.State).To(Equal(storage.State{
					IAAS: "aws",
					KeyPair: storage.KeyPair{
						Name: "some-aws-keypair",
					},
				}))

				Expect(state).To(Equal(storage.State{
					IAAS: "aws",
					KeyPair: storage.KeyPair{
						Name: "some-new-aws-keypair",
					},
				}))
			})

			Context("failure cases", func() {
				It("returns an error when rotate fails", func() {
					awsManager.RotateCall.Returns.Error = errors.New("call to rotate failed")
					_, err := keyPairManager.Rotate(storage.State{
						IAAS: "aws",
					})
					Expect(err).To(MatchError("call to rotate failed"))
				})
			})
		})

		Context("when iaas is gcp", func() {
			It("calls the gcp manager rotate and returns state", func() {
				state, err := keyPairManager.Rotate(storage.State{
					IAAS: "gcp",
					KeyPair: storage.KeyPair{
						Name: "some-gcp-keypair",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(gcpManager.RotateCall.CallCount).To(Equal(1))
				Expect(gcpManager.RotateCall.Receives.State).To(Equal(storage.State{
					IAAS: "gcp",
					KeyPair: storage.KeyPair{
						Name: "some-gcp-keypair",
					},
				}))

				Expect(state).To(Equal(storage.State{
					IAAS: "gcp",
					KeyPair: storage.KeyPair{
						Name: "some-new-gcp-keypair",
					},
				}))
			})

			Context("failure cases", func() {
				It("returns an error when rotate fails", func() {
					gcpManager.RotateCall.Returns.Error = errors.New("call to rotate failed")
					_, err := keyPairManager.Rotate(storage.State{
						IAAS: "gcp",
					})
					Expect(err).To(MatchError("call to rotate failed"))
				})
			})
		})

		Context("failure cases", func() {
			Context("when iaas is invalid", func() {
				It("returns an error", func() {
					_, err := keyPairManager.Rotate(storage.State{
						IAAS: "invalid-iaas",
					})
					Expect(err).To(MatchError("invalid iaas was provided: invalid-iaas"))
				})
			})
		})
	})
})
