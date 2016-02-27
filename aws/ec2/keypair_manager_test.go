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
			validator    *fakes.KeyPairValidator
			generator    *fakes.KeyPairGenerator
			uploader     *fakes.KeyPairUploader
			retriever    *fakes.KeyPairRetriever
			verifier     *fakes.KeyPairVerifier
			ec2Session   *fakes.EC2Session
			manager      ec2.KeyPairManager
		)

		BeforeEach(func() {
			validator = &fakes.KeyPairValidator{}
			generator = &fakes.KeyPairGenerator{}
			uploader = &fakes.KeyPairUploader{}
			retriever = &fakes.KeyPairRetriever{}
			verifier = &fakes.KeyPairVerifier{}
			ec2Session = &fakes.EC2Session{}
			manager = ec2.NewKeyPairManager(validator, generator, uploader, retriever, verifier)
		})

		Context("no keypair in state file", func() {
			BeforeEach(func() {
				stateKeyPair = ec2.KeyPair{}
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

			It("generates and uploads a keypair", func() {
				generator.GenerateCall.Returns.KeyPair = ec2.KeyPair{
					Name:       "my-keypair",
					PublicKey:  []byte("public"),
					PrivateKey: []byte("private"),
				}

				keypair, err := manager.Sync(ec2Session, stateKeyPair)
				Expect(err).NotTo(HaveOccurred())
				Expect(keypair).To(Equal(ec2.KeyPair{
					Name:       "my-keypair",
					PublicKey:  []byte("public"),
					PrivateKey: []byte("private"),
				}))

				Expect(validator.ValidateCall.CallCount).To(Equal(0))
				Expect(generator.GenerateCall.CallCount).To(Equal(1))
				Expect(uploader.UploadCall.Receives.Session).To(Equal(ec2Session))
				Expect(uploader.UploadCall.Receives.KeyPair).To(Equal(ec2.KeyPair{
					Name:       "my-keypair",
					PublicKey:  []byte("public"),
					PrivateKey: []byte("private"),
				}))
				Expect(retriever.RetrieveCall.CallCount).To(Equal(2))
				Expect(verifier.VerifyCall.Receives.Fingerprint).To(Equal("fingerprint"))
				Expect(verifier.VerifyCall.Receives.PEMData).To(Equal([]byte("private")))
			})

			Context("error cases", func() {
				Context("when the keypair cannot be generated", func() {
					It("returns an error", func() {
						generator.GenerateCall.Returns.Error = errors.New("failed to generate")

						_, err := manager.Sync(ec2Session, stateKeyPair)
						Expect(err).To(MatchError("failed to generate"))
					})
				})

				Context("when the keypair cannot be uploaded", func() {
					It("returns an error", func() {
						uploader.UploadCall.Returns.Error = errors.New("failed to upload")

						_, err := manager.Sync(ec2Session, stateKeyPair)
						Expect(err).To(MatchError("failed to upload"))
					})
				})

				Context("remote keypair retrieve fails", func() {
					It("returns an error", func() {
						retriever.RetrieveCall.Stub = nil
						retriever.RetrieveCall.Returns.Error = errors.New("keypair retrieve failed")

						_, err := manager.Sync(ec2Session, stateKeyPair)
						Expect(err).To(MatchError("keypair retrieve failed"))
					})
				})

				Context("when keypair retrieve fails after the first time", func() {
					It("returns an error", func() {
						retriever.RetrieveCall.Stub = func(_ ec2.Session, name string) (ec2.KeyPairInfo, bool, error) {
							if retriever.RetrieveCall.CallCount == 1 {
								return ec2.KeyPairInfo{}, false, nil
							}

							return ec2.KeyPairInfo{}, false, errors.New("failed to retrieve")
						}

						_, err := manager.Sync(ec2Session, stateKeyPair)
						Expect(err).To(MatchError("failed to retrieve"))
					})
				})

				Context("when fingerprint cannot be verified", func() {
					It("returns an error", func() {
						verifier.VerifyCall.Returns.Error = errors.New("fingerprint not verified")

						_, err := manager.Sync(ec2Session, stateKeyPair)
						Expect(err).To(MatchError("fingerprint not verified"))
					})
				})

				Context("no remote key found after first retrieve", func() {
					It("returns an error", func() {
						retriever.RetrieveCall.Stub = nil
						retriever.RetrieveCall.Returns.Present = false

						_, err := manager.Sync(ec2Session, stateKeyPair)
						Expect(err).To(MatchError("could not retrieve keypair for verification"))
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

			It("uploads the given keypair", func() {
				keypair, err := manager.Sync(ec2Session, stateKeyPair)
				Expect(err).NotTo(HaveOccurred())
				Expect(keypair).To(Equal(ec2.KeyPair{
					Name:       "my-keypair",
					PublicKey:  []byte("public"),
					PrivateKey: []byte("private"),
				}))

				Expect(validator.ValidateCall.CallCount).To(Equal(1))
				Expect(validator.ValidateCall.Receives.PEMData).To(Equal([]byte("private")))
				Expect(uploader.UploadCall.Receives.Session).To(Equal(ec2Session))
				Expect(uploader.UploadCall.Receives.KeyPair).To(Equal(ec2.KeyPair{
					Name:       "my-keypair",
					PublicKey:  []byte("public"),
					PrivateKey: []byte("private"),
				}))
				Expect(retriever.RetrieveCall.CallCount).To(Equal(2))
				Expect(verifier.VerifyCall.Receives.Fingerprint).To(Equal("fingerprint"))
				Expect(verifier.VerifyCall.Receives.PEMData).To(Equal([]byte("private")))
			})

			Context("failure cases", func() {
				Context("when the keypair is not valid", func() {
					It("returns an error", func() {
						validator.ValidateCall.Returns.Error = errors.New("failed to validate")

						_, err := manager.Sync(ec2Session, stateKeyPair)
						Expect(err).To(MatchError("failed to validate"))
					})
				})

				It("returns an error when key upload fails", func() {
					uploader.UploadCall.Returns.Error = errors.New("upload failed")
					retriever.RetrieveCall.Returns.Present = false

					_, err := manager.Sync(ec2Session, stateKeyPair)
					Expect(err).NotTo(BeNil())
					Expect(err).To(MatchError("upload failed"))
				})

				Context("remote keypair retrieve fails", func() {
					It("returns an error", func() {
						retriever.RetrieveCall.Stub = nil
						retriever.RetrieveCall.Returns.Error = errors.New("keypair retrieve failed")

						_, err := manager.Sync(ec2Session, ec2.KeyPair{})
						Expect(err).To(MatchError("keypair retrieve failed"))
					})
				})

				Context("when keypair retrieve fails after the first time", func() {
					It("returns an error", func() {
						retriever.RetrieveCall.Stub = func(_ ec2.Session, name string) (ec2.KeyPairInfo, bool, error) {
							if retriever.RetrieveCall.CallCount == 1 {
								return ec2.KeyPairInfo{}, false, nil
							}

							return ec2.KeyPairInfo{}, false, errors.New("failed to retrieve")
						}

						_, err := manager.Sync(ec2Session, stateKeyPair)
						Expect(err).To(MatchError("failed to retrieve"))
					})
				})

				Context("when fingerprint cannot be verified", func() {
					It("returns an error", func() {
						verifier.VerifyCall.Returns.Error = errors.New("fingerprint not verified")

						_, err := manager.Sync(ec2Session, stateKeyPair)
						Expect(err).To(MatchError("fingerprint not verified"))
					})
				})

				Context("no remote key found after first retrieve", func() {
					It("returns an error", func() {
						retriever.RetrieveCall.Stub = func(ec2.Session, string) (ec2.KeyPairInfo, bool, error) {
							if retriever.RetrieveCall.CallCount == 1 {
								return ec2.KeyPairInfo{}, true, nil
							}

							return ec2.KeyPairInfo{}, false, nil
						}

						_, err := manager.Sync(ec2Session, stateKeyPair)
						Expect(err).To(MatchError("could not retrieve keypair for verification"))
					})
				})
			})
		})
	})
})
