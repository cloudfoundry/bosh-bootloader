package boshinit_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/boshinit"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/ssl"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeployInput", func() {
	var (
		fakeStringGenerator         *fakes.StringGenerator
		state                       storage.State
		infrastructureConfiguration boshinit.InfrastructureConfiguration
		envID                       string
	)

	Describe("NewDeployInput", func() {
		BeforeEach(func() {
			fakeStringGenerator = &fakes.StringGenerator{}
			state = storage.State{
				KeyPair: storage.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
				BOSH: storage.BOSH{
					DirectorSSLCertificate: "some-ssl-cert",
					DirectorSSLPrivateKey:  "some-ssl-private-key",
					Credentials: map[string]string{
						"some-user": "some-password",
					},
					State: map[string]interface{}{
						"some-state-key": "some-state-value",
					},
					DirectorName:     "some-director-name",
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
				},
			}

			infrastructureConfiguration = boshinit.InfrastructureConfiguration{
				AWSRegion:        "some-aws-region",
				SubnetID:         "some-subnet-id",
				AvailabilityZone: "some-az",
				ElasticIP:        "some-eip",
				AccessKeyID:      "some-access-key-id",
				SecretAccessKey:  "some-secret-access-key",
				SecurityGroup:    "some-security-group",
			}

			envID = "some-env-id"
		})

		It("constructs a DeployInput given a state", func() {
			deployInput, err := boshinit.NewDeployInput(state, infrastructureConfiguration, fakeStringGenerator, envID)

			Expect(err).NotTo(HaveOccurred())
			Expect(deployInput).To(Equal(boshinit.DeployInput{
				DirectorName:     "some-director-name",
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
				State: map[string]interface{}{
					"some-state-key": "some-state-value",
				},
				InfrastructureConfiguration: boshinit.InfrastructureConfiguration{
					AWSRegion:        "some-aws-region",
					SubnetID:         "some-subnet-id",
					AvailabilityZone: "some-az",
					ElasticIP:        "some-eip",
					AccessKeyID:      "some-access-key-id",
					SecretAccessKey:  "some-secret-access-key",
					SecurityGroup:    "some-security-group",
				},
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

		Context("when existing state contains bosh state without director name", func() {
			It("sets director name to my-bosh", func() {
				state.BOSH.DirectorName = ""
				deployInput, err := boshinit.NewDeployInput(state, infrastructureConfiguration, fakeStringGenerator, envID)

				Expect(err).NotTo(HaveOccurred())
				Expect(deployInput.DirectorName).To(Equal("my-bosh"))
			})
		})

		It("does not modify the struct references in the state", func() {
			state := storage.State{
				AWS: storage.AWS{
					Region: "some-aws-region",
				},
				KeyPair: storage.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
				BOSH: storage.BOSH{
					DirectorSSLCertificate: "some-ssl-cert",
					DirectorSSLPrivateKey:  "some-ssl-private-key",
					Credentials: map[string]string{
						"some-user": "some-password",
					},
					State: map[string]interface{}{
						"some-state-key": "some-state-value",
					},
					DirectorName:     "some-director-name",
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
				},
			}

			_, err := boshinit.NewDeployInput(state, boshinit.InfrastructureConfiguration{}, fakeStringGenerator, envID)
			Expect(err).NotTo(HaveOccurred())

			Expect(state).To(Equal(storage.State{
				AWS: storage.AWS{
					Region: "some-aws-region",
				},
				KeyPair: storage.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
				BOSH: storage.BOSH{
					DirectorSSLCertificate: "some-ssl-cert",
					DirectorSSLPrivateKey:  "some-ssl-private-key",
					Credentials: map[string]string{
						"some-user": "some-password",
					},
					State: map[string]interface{}{
						"some-state-key": "some-state-value",
					},
					DirectorName:     "some-director-name",
					DirectorUsername: "some-director-username",
					DirectorPassword: "some-director-password",
				},
			}))
		})

		It("handles empty state, by assigning director name and generating credentials if they don't exist", func() {
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
			deployInput, err := boshinit.NewDeployInput(storage.State{}, boshinit.InfrastructureConfiguration{}, fakeStringGenerator, envID)

			Expect(err).NotTo(HaveOccurred())
			Expect(deployInput).To(Equal(boshinit.DeployInput{
				State:            map[string]interface{}{},
				DirectorName:     "bosh-some-env-id",
				DirectorUsername: "some-generated-username",
				DirectorPassword: "some-generated-password",
			}))
			Expect(fakeStringGenerator.GenerateCall.Receives.Prefixes).To(Equal([]string{"user-", "p-"}))
			Expect(fakeStringGenerator.GenerateCall.Receives.Lengths).To(Equal([]int{7, 15}))
		})

		Describe("failure cases", func() {
			It("returns an error when director username generation fails", func() {
				fakeStringGenerator.GenerateCall.Returns.Error = errors.New("failed to generate username")
				_, err := boshinit.NewDeployInput(storage.State{}, boshinit.InfrastructureConfiguration{}, fakeStringGenerator, "")

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
				_, err := boshinit.NewDeployInput(storage.State{}, boshinit.InfrastructureConfiguration{}, fakeStringGenerator, "")
				Expect(err).To(MatchError("failed to generate password"))
			})
		})
	})
})
