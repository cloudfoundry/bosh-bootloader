package boshinit_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BOSHDeployInput", func() {
	var (
		fakeStringGenerator *fakes.StringGenerator
	)

	Describe("NewBOSHDeployInput", func() {
		BeforeEach(func() {
			fakeStringGenerator = &fakes.StringGenerator{}
		})

		It("constructs a BOSHDeployInput given a state", func() {
			state := storage.State{
				AWS: storage.AWS{
					Region: "some-aws-region",
				},
				KeyPair: &storage.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
				BOSH: &storage.BOSH{
					DirectorSSLCertificate: "some-ssl-cert",
					DirectorSSLPrivateKey:  "some-ssl-private-key",
					Credentials: map[string]string{
						"some-user": "some-password",
					},
					State: map[string]interface{}{
						"some-state-key": "some-state-value",
					},
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
				},
			}

			stack := cloudformation.Stack{
				Outputs: map[string]string{
					"some-stack-output-key": "some-stack-output-value",
				},
			}

			boshDeployInput, err := boshinit.NewBOSHDeployInput(state, stack, fakeStringGenerator)

			Expect(err).NotTo(HaveOccurred())
			Expect(boshDeployInput).To(Equal(boshinit.BOSHDeployInput{
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
				State: map[string]interface{}{
					"some-state-key": "some-state-value",
				},
				Stack: cloudformation.Stack{
					Outputs: map[string]string{
						"some-stack-output-key": "some-stack-output-value",
					},
				},
				AWSRegion: "some-aws-region",
				SSLKeyPair: ssl.KeyPair{
					Certificate: []byte("some-ssl-cert"),
					PrivateKey:  []byte("some-ssl-private-key"),
				},
				EC2KeyPair: ec2.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
				Credentials: map[string]string{
					"some-user": "some-password",
				},
			}))
		})

		It("does not modify the struct references in the state", func() {
			state := storage.State{
				AWS: storage.AWS{
					Region: "some-aws-region",
				},
				KeyPair: &storage.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
				BOSH: &storage.BOSH{
					DirectorSSLCertificate: "some-ssl-cert",
					DirectorSSLPrivateKey:  "some-ssl-private-key",
					Credentials: map[string]string{
						"some-user": "some-password",
					},
					State: map[string]interface{}{
						"some-state-key": "some-state-value",
					},
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
				},
			}

			stack := cloudformation.Stack{
				Outputs: map[string]string{
					"some-stack-output-key": "some-stack-output-value",
				},
			}

			_, err := boshinit.NewBOSHDeployInput(state, stack, fakeStringGenerator)
			Expect(err).NotTo(HaveOccurred())

			Expect(state).To(Equal(storage.State{
				AWS: storage.AWS{
					Region: "some-aws-region",
				},
				KeyPair: &storage.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
				BOSH: &storage.BOSH{
					DirectorSSLCertificate: "some-ssl-cert",
					DirectorSSLPrivateKey:  "some-ssl-private-key",
					Credentials: map[string]string{
						"some-user": "some-password",
					},
					State: map[string]interface{}{
						"some-state-key": "some-state-value",
					},
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
				},
			}))
		})

		It("handles empty state, generating director credentials if they don't exist", func() {
			fakeStringGenerator.GenerateCall.Stub = func(prefix string, length int) (string, error) {
				switch fakeStringGenerator.GenerateCall.CallCount {
				case 0:
					return "some-generated-username", nil
				case 1:
					return "some-generated-password", nil
				default:
					return "", errors.New("too many calls to password generator")
				}
			}
			boshDeployInput, err := boshinit.NewBOSHDeployInput(storage.State{}, cloudformation.Stack{}, fakeStringGenerator)

			Expect(err).NotTo(HaveOccurred())
			Expect(boshDeployInput).To(Equal(boshinit.BOSHDeployInput{
				State:            map[string]interface{}{},
				DirectorUsername: "some-generated-username",
				DirectorPassword: "some-generated-password",
			}))
		})

		Describe("failure cases", func() {
			It("returns an error when director username generation fails", func() {
				fakeStringGenerator.GenerateCall.Returns.Error = errors.New("failed to generate username")
				_, err := boshinit.NewBOSHDeployInput(storage.State{}, cloudformation.Stack{}, fakeStringGenerator)

				Expect(err).To(MatchError("failed to generate username"))
			})

			It("returns an error when director username generation fails", func() {
				fakeStringGenerator.GenerateCall.Stub = func(prefix string, length int) (string, error) {
					switch fakeStringGenerator.GenerateCall.CallCount {
					case 0:
						return "", nil
					default:
						return "", errors.New("failed to generate password")
					}
				}
				_, err := boshinit.NewBOSHDeployInput(storage.State{}, cloudformation.Stack{}, fakeStringGenerator)
				Expect(err).To(MatchError("failed to generate password"))
			})
		})
	})
})
