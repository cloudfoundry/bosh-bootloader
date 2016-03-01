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
			retriever    *fakes.KeyPairRetriever
			ec2Client    *fakes.EC2Client
			manager      ec2.KeyPairManager
		)

		BeforeEach(func() {
			creator = &fakes.KeyPairCreator{}
			retriever = &fakes.KeyPairRetriever{}
			ec2Client = &fakes.EC2Client{}
			manager = ec2.NewKeyPairManager(creator, retriever)
		})

		Context("no keypair in state file", func() {
			BeforeEach(func() {
				stateKeyPair = ec2.KeyPair{}
				retriever.RetrieveCall.Returns.Present = false
			})

			It("creates a keypair", func() {
				creator.CreateCall.Returns.KeyPair = ec2.KeyPair{
					Name:       "my-keypair",
					PublicKey:  []byte("public"),
					PrivateKey: []byte("private"),
				}

				keypair, err := manager.Sync(ec2Client, stateKeyPair)
				Expect(err).NotTo(HaveOccurred())
				Expect(keypair).To(Equal(ec2.KeyPair{
					Name:       "my-keypair",
					PublicKey:  []byte("public"),
					PrivateKey: []byte("private"),
				}))

				Expect(creator.CreateCall.Receives.Session).To(Equal(ec2Client))
				Expect(retriever.RetrieveCall.CallCount).To(Equal(1))
			})

			Context("error cases", func() {
				Context("when the keypair cannot be created", func() {
					It("returns an error", func() {
						creator.CreateCall.Returns.Error = errors.New("failed to create key pair")

						_, err := manager.Sync(ec2Client, stateKeyPair)
						Expect(err).To(MatchError("failed to create key pair"))
					})
				})

				Context("remote keypair retrieve fails", func() {
					It("returns an error", func() {
						retriever.RetrieveCall.Stub = nil
						retriever.RetrieveCall.Returns.Error = errors.New("keypair retrieve failed")

						_, err := manager.Sync(ec2Client, stateKeyPair)
						Expect(err).To(MatchError("keypair retrieve failed"))
					})
				})
			})
		})

		Context("when the keypair is in the state file, but not on ec2", func() {
			BeforeEach(func() {
				stateKeyPair = ec2.KeyPair{
					Name:       "my-keypair",
					PublicKey:  []byte("public"),
					PrivateKey: []byte("private"),
				}
				retriever.RetrieveCall.Stub = func(_ ec2.Session, name string) (ec2.KeyPairInfo, bool, error) {
					if retriever.RetrieveCall.CallCount == 1 {
						return ec2.KeyPairInfo{}, false, nil
					}

					return ec2.KeyPairInfo{
						Name:        name,
						Fingerprint: "fingerprint",
					}, true, nil
				}
			})

			It("creates a keypair", func() {
				creator.CreateCall.Returns.KeyPair = ec2.KeyPair{
					Name:       "my-keypair",
					PublicKey:  []byte("public"),
					PrivateKey: []byte("private"),
				}

				keypair, err := manager.Sync(ec2Client, stateKeyPair)
				Expect(err).NotTo(HaveOccurred())
				Expect(keypair).To(Equal(ec2.KeyPair{
					Name:       "my-keypair",
					PublicKey:  []byte("public"),
					PrivateKey: []byte("private"),
				}))

				Expect(creator.CreateCall.Receives.Session).To(Equal(ec2Client))
				Expect(retriever.RetrieveCall.CallCount).To(Equal(1))
			})

			Context("failure cases", func() {
				Context("when the keypair cannot be created", func() {
					It("returns an error", func() {
						creator.CreateCall.Returns.Error = errors.New("failed to create key pair")

						_, err := manager.Sync(ec2Client, stateKeyPair)
						Expect(err).To(MatchError("failed to create key pair"))
					})
				})

				Context("remote keypair retrieve fails", func() {
					It("returns an error", func() {
						retriever.RetrieveCall.Stub = nil
						retriever.RetrieveCall.Returns.Error = errors.New("keypair retrieve failed")

						_, err := manager.Sync(ec2Client, ec2.KeyPair{})
						Expect(err).To(MatchError("keypair retrieve failed"))
					})
				})
			})
		})
	})
})
