package ec2_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KeyPairManager", func() {
	Describe("Sync", func() {
		var (
			stateKeyPair ec2.KeyPair
			creator      *fakes.KeyPairCreator
			checker      *fakes.KeyPairChecker
			logger       *fakes.Logger
			manager      ec2.KeyPairManager
		)

		BeforeEach(func() {
			creator = &fakes.KeyPairCreator{}
			checker = &fakes.KeyPairChecker{}
			logger = &fakes.Logger{}
			manager = ec2.NewKeyPairManager(creator, checker, logger)
		})

		Context("no keypair in state file", func() {
			BeforeEach(func() {
				stateKeyPair = ec2.KeyPair{Name: "keypair-some-env-id"}
				checker.HasKeyPairCall.Returns.Present = false
			})

			It("creates a keypair", func() {
				creator.CreateCall.Returns.KeyPair = ec2.KeyPair{
					Name:       "keypair-some-env-id",
					PublicKey:  "public",
					PrivateKey: "private",
				}

				keypair, err := manager.Sync(stateKeyPair)
				Expect(err).NotTo(HaveOccurred())
				Expect(keypair).To(Equal(ec2.KeyPair{
					Name:       "keypair-some-env-id",
					PublicKey:  "public",
					PrivateKey: "private",
				}))

				Expect(creator.CreateCall.Receives.KeyPairName).To(Equal("keypair-some-env-id"))

				Expect(checker.HasKeyPairCall.CallCount).To(Equal(1))
				Expect(logger.StepCall.Receives.Message).To(Equal("creating keypair: `%s`"))
				Expect(logger.StepCall.Receives.Arguments[0]).To(Equal("keypair-some-env-id"))
			})

			Context("error cases", func() {
				Context("when the keypair cannot be created", func() {
					It("returns an error", func() {
						creator.CreateCall.Returns.Error = errors.New("failed to create key pair")

						_, err := manager.Sync(stateKeyPair)
						Expect(err).To(MatchError("failed to create key pair"))
					})
				})

				Context("when remote keypair retrieve fails", func() {
					It("returns an error", func() {
						checker.HasKeyPairCall.Stub = nil
						checker.HasKeyPairCall.Returns.Error = errors.New("keypair retrieve failed")

						_, err := manager.Sync(stateKeyPair)
						Expect(err).To(MatchError("keypair retrieve failed"))
					})
				})
			})
		})

		Context("when the keypair is in the state file, but not on ec2", func() {
			BeforeEach(func() {
				stateKeyPair = ec2.KeyPair{
					Name:       "my-keypair",
					PublicKey:  "public",
					PrivateKey: "private",
				}
				checker.HasKeyPairCall.Stub = func(name string) (bool, error) {
					if checker.HasKeyPairCall.CallCount == 1 {
						return false, nil
					}

					return true, nil
				}
			})

			It("creates a keypair", func() {
				creator.CreateCall.Returns.KeyPair = ec2.KeyPair{
					Name:       "my-keypair",
					PublicKey:  "public",
					PrivateKey: "private",
				}

				keypair, err := manager.Sync(stateKeyPair)
				Expect(err).NotTo(HaveOccurred())
				Expect(keypair).To(Equal(ec2.KeyPair{
					Name:       "my-keypair",
					PublicKey:  "public",
					PrivateKey: "private",
				}))

				Expect(checker.HasKeyPairCall.CallCount).To(Equal(1))
			})

			Context("failure cases", func() {
				Context("when the keypair cannot be created", func() {
					It("returns an error", func() {
						creator.CreateCall.Returns.Error = errors.New("failed to create key pair")

						_, err := manager.Sync(stateKeyPair)
						Expect(err).To(MatchError("failed to create key pair"))
					})
				})

				Context("remote keypair retrieve fails", func() {
					It("returns an error", func() {
						checker.HasKeyPairCall.Stub = nil
						checker.HasKeyPairCall.Returns.Error = errors.New("keypair retrieve failed")

						_, err := manager.Sync(ec2.KeyPair{})
						Expect(err).To(MatchError("keypair retrieve failed"))
					})
				})
			})
		})

		Context("when the keypair is in the state file and on ec2", func() {
			BeforeEach(func() {
				stateKeyPair = ec2.KeyPair{
					Name:       "my-keypair",
					PublicKey:  "public",
					PrivateKey: "private",
				}
				checker.HasKeyPairCall.Returns.Present = true
			})

			It("logs that the existing keypair will be used", func() {
				_, err := manager.Sync(stateKeyPair)
				Expect(err).NotTo(HaveOccurred())
				Expect(logger.StepCall.Receives.Message).To(Equal("using existing keypair"))
			})
		})
	})
})
