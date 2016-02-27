package unsupported_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	privateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAvXMDyDguAnAQqQF0mAff2etww+G+Mpf07hUlnGDkzwELdQZ+
dcYCPOWLISUuGH3EJVrlOtaxc/UY83P/Czg85sGnSkccLrApUH3aQFdx0QQ3D7ZD
qrmxkvLHKhf11F7Mz4o3/6bv/M2SqPtWGO5NSdtp77QX+my00KuloMyTv+TNZ8A+
DTJPW47Ip0Kq/NUK0rL6kUS7zn1KAkXlUST1GGR1vISn0OOP0PfD0szQZaYCmAJ6
DOzGBd2/4abscTP7fFWxFpEGWRZEhJ+YnxKhF6myQ9Bs6qPXNJPA1OyN1LXtC0Uh
+x6ByAQLrix0N71JJ7grg1y28bcag3kfvgHwCwIDAQABAoIBAEB7vIbS8H4t7M3J
zAjPbVc8d0aFOPr5lAnRstqWdGstPNwZWMP3oN1feErQ3+7AKBpa5PlxCDei7lo3
WlFUVA5rTejPaX1OwtE99SK/YOM3HxK/BCtBR3rwHfBq9WbS2b2umz7ucHNI+amA
2x5jRnVkNJu9XggEJkt8kUS5PXUr8oMjOdMtNCFum4roicFkuvOBHLWlPmX9UvbY
zHvNw1KfygrZq9NvYiDjOES234VYZb7L1AcC/rDoAC353EE4qt3Xm0F+OE2ai7dx
u3JfdCwIzdRLw5k5CyFgRiC+8i6/X3lNJR86Rvd1zC9r2mEwLt2rZl8kg/2pcBvb
HzxDzVECgYEAyznER78IDiqoYe0mfQJ+OlLv6+ZZj/pVjLhQtLwSk66dQLuUFBhz
LhEXXHMuyh9Y0Utw2OiawlYlkCuw/3qH1dNCbE1YxBSpsr9HIjHLBhLaYFLBqOQS
X973jLG/JDUuiZnrZoQ/ScNwaSY8b9gh/CAHzBVjIA4rSzI68QWgTLkCgYEA7qVo
pHDbenYmh9pYiXGym9fu2RxlYYX1n5BDiLxXMMLVEvsXFV23kuYoQS5wqASW8v4T
kLLDT6AsOK9541WGN5p42uTQigTEaouXklMI5Krk2d1kYSh+/EI18dytmk8NHkeD
ObW+3Xk5Q4RSdl3DlPET59F2rbxoRUcs9KVpKOMCgYB+A/3/9ybJkg4DWwhor+kR
xWfcQWP78WCm94uj5pMmXDpKb4Ysx9R0FkkEHLBAyRtL/JmnBuUf6Ec2lMEWSiZ8
opknival76IiopU7UODxjTM4U1ien339UMbzySwbCZcn3/emBA8ycCv+J6WGPOEl
876iAAkNUXvrDuSZm8GAkQKBgQC6v7vseth1s4GhbA8+tzeS1t51DdCUCXVVwVnn
5aLBaKW+7bh5otXl4a/8me/Uu4q4anU7FXjblbclQMQ8Tw/x8TLD8Kz0ZJij28rn
2YyrDMR7bNGBamQ82T9Hnm5Hw7a7TDD3dy7+Nz/FgwXY1LUZl7IBBZw+hqJ+HB2k
8NAjCwKBgFFrbUPiQfqDoftkhoofEAqTbz8OBEs4iJwcpNxr9Q9Yz/Bbq/2Kb4Tz
Tur818l/ASHdiwFbYsGz7CJfGgmjNIsL4s2QhTkJgqq0f+Nv8NVNkJ2PFtYZS1EG
O0flBWBP4MtvJjixn7G49dynF6PTYJPZBEjcL1R91qkaPYEAOcl4
-----END RSA PRIVATE KEY-----`
)

var _ = Describe("CreateBoshAWSKeyPair", func() {
	var (
		command          unsupported.CreateBoshAWSKeyPair
		keypairGenerator *fakes.KeyPairGenerator
		keypairRetriever *fakes.KeyPairRetriever
		keypairUploader  *fakes.KeyPairUploader
		session          *fakes.EC2Session
		sessionProvider  *fakes.EC2SessionProvider
	)

	BeforeEach(func() {
		keypairGenerator = &fakes.KeyPairGenerator{}
		keypairUploader = &fakes.KeyPairUploader{}
		keypairRetriever = &fakes.KeyPairRetriever{}

		session = &fakes.EC2Session{}
		sessionProvider = &fakes.EC2SessionProvider{}
		sessionProvider.SessionCall.Returns.Session = session

		command = unsupported.NewCreateBoshAWSKeyPair(keypairRetriever, keypairGenerator, keypairUploader, sessionProvider)
	})

	Describe("Execute", func() {
		var incomingState storage.State

		BeforeEach(func() {
			incomingState = storage.State{
				AWS: storage.AWS{
					AccessKeyID:     "some-aws-access-key-id",
					SecretAccessKey: "some-aws-secret-access-key",
					Region:          "some-aws-region",
				},
			}
		})

		It("generates a new keypair", func() {
			_, err := command.Execute(commands.GlobalFlags{
				EndpointOverride: "some-endpoint-override",
				StateDir:         "/some/state/dir",
			}, incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(keypairGenerator.GenerateCall.CallCount).To(Equal(1))
		})

		It("initializes a new session with the correct config", func() {
			_, err := command.Execute(commands.GlobalFlags{
				EndpointOverride: "some-endpoint-override",
			}, incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(sessionProvider.SessionCall.Receives.Config).To(Equal(aws.Config{
				AccessKeyID:      "some-aws-access-key-id",
				SecretAccessKey:  "some-aws-secret-access-key",
				Region:           "some-aws-region",
				EndpointOverride: "some-endpoint-override",
			}))
		})

		It("uploads the generated keypair", func() {
			keypairGenerator.GenerateCall.Returns.KeyPair = ec2.KeyPair{
				Name:      "some-name",
				PublicKey: []byte("some-key"),
			}

			_, err := command.Execute(commands.GlobalFlags{
				EndpointOverride: "some-endpoint-override",
				StateDir:         "/some/state/dir",
			}, incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(keypairUploader.UploadCall.Receives.Session).To(Equal(session))
			Expect(keypairUploader.UploadCall.Receives.KeyPair).To(Equal(ec2.KeyPair{
				Name:      "some-name",
				PublicKey: []byte("some-key"),
			}))
		})

		It("returns a state with keypair and name", func() {
			keypairGenerator.GenerateCall.Returns.KeyPair = ec2.KeyPair{
				Name:       "some-name",
				PublicKey:  []byte("some-public-key"),
				PrivateKey: []byte("some-private-key"),
			}

			state, err := command.Execute(commands.GlobalFlags{
				EndpointOverride: "some-endpoint-override",
				StateDir:         "/some/state/dir",
			}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(state).To(Equal(storage.State{
				AWS: storage.AWS{
					AccessKeyID:     "some-aws-access-key-id",
					SecretAccessKey: "some-aws-secret-access-key",
					Region:          "some-aws-region",
				},
				KeyPair: &storage.KeyPair{
					Name:       "some-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
			}))
		})

		Context("idempotently generates a keypair", func() {
			It("uses the keypair in the store if it matches the keypair in AWS", func() {
				incomingState = storage.State{
					KeyPair: &storage.KeyPair{
						Name:       "some-name",
						PrivateKey: privateKey,
					},
				}

				keypairRetriever.RetrieveCall.Returns.KeyPairInfo = ec2.KeyPairInfo{
					Name:        "some-name",
					Fingerprint: "5d:f1:5a:6b:22:87:27:a5:e3:33:5e:d2:c9:7f:2e:08",
				}

				_, err := command.Execute(commands.GlobalFlags{
					EndpointOverride: "some-endpoint-override",
					StateDir:         "/some/state/dir",
				}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(keypairGenerator.GenerateCall.CallCount).To(Equal(0))

				Expect(keypairRetriever.RetrieveCall.CallCount).To(Equal(1))
				Expect(keypairRetriever.RetrieveCall.Recieves.Session).To(Equal(session))
				Expect(keypairRetriever.RetrieveCall.Recieves.Name).To(Equal("some-name"))
			})

			It("uploads the keypair in the store if the keypair does not exist on AWS", func() {
				incomingState = storage.State{
					KeyPair: &storage.KeyPair{
						Name:      "some-name",
						PublicKey: "some-public-key",
					},
				}

				keypairRetriever.RetrieveCall.Returns.Present = false

				_, err := command.Execute(commands.GlobalFlags{
					EndpointOverride: "some-endpoint-override",
					StateDir:         "/some/state/dir",
				}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(keypairRetriever.RetrieveCall.CallCount).To(Equal(1))
				Expect(keypairRetriever.RetrieveCall.Recieves.Session).To(Equal(session))
				Expect(keypairRetriever.RetrieveCall.Recieves.Name).To(Equal("some-name"))

				Expect(keypairGenerator.GenerateCall.CallCount).To(Equal(0))

				Expect(keypairUploader.UploadCall.Receives.Session).To(Equal(session))
				Expect(keypairUploader.UploadCall.Receives.KeyPair).To(Equal(ec2.KeyPair{
					Name:      "some-name",
					PublicKey: []byte("some-public-key"),
				}))
			})

			Context("failure cases", func() {
				It("returns an error when the fingerprints don't match", func() {
					keypairGenerator.GenerateCall.Returns.KeyPair = ec2.KeyPair{
						Name:      "some-new-key-name",
						PublicKey: []byte("some-new-public-key"),
					}

					incomingState = storage.State{
						KeyPair: &storage.KeyPair{
							Name:       "some-name",
							PrivateKey: privateKey,
						},
					}

					keypairRetriever.RetrieveCall.Returns.Present = true
					keypairRetriever.RetrieveCall.Returns.KeyPairInfo = ec2.KeyPairInfo{
						Name:        "some-name",
						Fingerprint: "some-fingerprint",
					}

					_, err := command.Execute(commands.GlobalFlags{
						EndpointOverride: "some-endpoint-override",
						StateDir:         "/some/state/dir",
					}, incomingState)

					Expect(err).To(MatchError("the local keypair fingerprint does not match the " +
						"keypair fingerprint on AWS, please open an issue at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new " +
						"if you require assistance."))

					Expect(keypairGenerator.GenerateCall.CallCount).To(Equal(0))
					Expect(keypairUploader.UploadCall.CallCount).To(Equal(0))
				})

				It("returns an error when the keypair can not be retrieved", func() {
					incomingState = storage.State{
						KeyPair: &storage.KeyPair{
							Name:       "some-name",
							PrivateKey: privateKey,
						},
					}
					keypairRetriever.RetrieveCall.Returns.Error = errors.New("something bad happened")

					_, err := command.Execute(commands.GlobalFlags{
						EndpointOverride: "some-endpoint-override",
						StateDir:         "/some/state/dir",
					}, incomingState)
					Expect(err).To(MatchError("something bad happened"))
				})

				It("returns an error when the private key is not in PEM format", func() {
					incomingState = storage.State{
						KeyPair: &storage.KeyPair{
							Name:       "some-name",
							PrivateKey: "some-private-key",
						},
					}

					keypairRetriever.RetrieveCall.Returns.Present = true
					keypairRetriever.RetrieveCall.Returns.KeyPairInfo = ec2.KeyPairInfo{
						Name:        "some-name",
						Fingerprint: "some-fingerprint",
					}

					_, err := command.Execute(commands.GlobalFlags{
						EndpointOverride: "some-endpoint-override",
						StateDir:         "/some/state/dir",
					}, incomingState)
					Expect(err).To(MatchError("the local keypair does not contain a valid PEM encoded private key, please open an " +
						"issue at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you require assistance."))
				})

				It("returns an error when the private key is not a valid rsa private key", func() {
					incomingState = storage.State{
						KeyPair: &storage.KeyPair{
							Name: "some-name",
							PrivateKey: `-----BEGIN RSA PRIVATE KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvXMDyDguAnAQqQF0mAff
2etww+G+Mpf07hUlnGDkzwELdQZ+dcYCPOWLISUuGH3EJVrlOtaxc/UY83P/Czg8
5sGnSkccLrApUH3aQFdx0QQ3D7ZDqrmxkvLHKhf11F7Mz4o3/6bv/M2SqPtWGO5N
Sdtp77QX+my00KuloMyTv+TNZ8A+DTJPW47Ip0Kq/NUK0rL6kUS7zn1KAkXlUST1
GGR1vISn0OOP0PfD0szQZaYCmAJ6DOzGBd2/4abscTP7fFWxFpEGWRZEhJ+YnxKh
F6myQ9Bs6qPXNJPA1OyN1LXtC0Uh+x6ByAQLrix0N71JJ7grg1y28bcag3kfvgHw
CwIDAQAB
-----END RSA PRIVATE KEY-----`,
						},
					}
					keypairRetriever.RetrieveCall.Returns.Present = true
					keypairRetriever.RetrieveCall.Returns.KeyPairInfo = ec2.KeyPairInfo{
						Name:        "some-name",
						Fingerprint: "some-fingerprint",
					}

					_, err := command.Execute(commands.GlobalFlags{
						EndpointOverride: "some-endpoint-override",
						StateDir:         "/some/state/dir",
					}, incomingState)
					Expect(err).To(MatchError("the local keypair does not contain a valid rsa private key, please open an issue " +
						"at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you require assistance."))
				})
			})
		})

		Context("failure cases", func() {
			It("returns an error when key generation fails", func() {
				keypairGenerator.GenerateCall.Returns.Error = errors.New("generate keys failed")

				_, err := command.Execute(commands.GlobalFlags{
					EndpointOverride: "some-endpoint-override",
					StateDir:         "/some/state/dir",
				}, incomingState)
				Expect(err).To(MatchError("generate keys failed"))
			})

			It("returns an error when key upload fails", func() {
				keypairUploader.UploadCall.Returns.Error = errors.New("upload keys failed")

				_, err := command.Execute(commands.GlobalFlags{
					EndpointOverride: "some-endpoint-override",
					StateDir:         "/some/state/dir",
				}, incomingState)
				Expect(err).To(MatchError("upload keys failed"))
			})

			It("returns an error when the session provided fails", func() {
				sessionProvider.SessionCall.Returns.Error = errors.New("failed to create session")

				_, err := command.Execute(commands.GlobalFlags{
					EndpointOverride: "some-endpoint-override",
					StateDir:         "/some/state/dir",
				}, incomingState)
				Expect(err).To(MatchError("failed to create session"))
			})
		})
	})
})
