package unsupported_test

import (
	"errors"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ec2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type FakeKeypairGenerator struct {
	GenerateCall struct {
		CallCount int
		Returns   struct {
			Keypair ec2.Keypair
			Error   error
		}
	}
}

func (g *FakeKeypairGenerator) Generate() (ec2.Keypair, error) {
	g.GenerateCall.CallCount++

	return g.GenerateCall.Returns.Keypair, g.GenerateCall.Returns.Error
}

type FakeKeypairUploader struct {
	UploadCall struct {
		Receives struct {
			Session ec2.Session
			Keypair ec2.Keypair
		}
		Returns struct {
			Error error
		}
	}
}

func (u *FakeKeypairUploader) Upload(session ec2.Session, keypair ec2.Keypair) error {
	u.UploadCall.Receives.Session = session
	u.UploadCall.Receives.Keypair = keypair

	return u.UploadCall.Returns.Error
}

type FakeSession struct{}

func (s *FakeSession) ImportKeyPair(*awsec2.ImportKeyPairInput) (*awsec2.ImportKeyPairOutput, error) {
	return nil, nil
}

type FakeSessionProvider struct {
	SessionCall struct {
		Receives struct {
			Config ec2.Config
		}
		Returns struct {
			Session ec2.Session
		}
	}
}

func (p *FakeSessionProvider) Session(config ec2.Config) ec2.Session {
	p.SessionCall.Receives.Config = config

	return p.SessionCall.Returns.Session
}

var _ = Describe("CreateBoshAWSKeypair", func() {
	var (
		command          unsupported.CreateBoshAWSKeypair
		keypairGenerator *FakeKeypairGenerator
		keypairUploader  *FakeKeypairUploader
		session          *FakeSession
		sessionProvider  *FakeSessionProvider
	)

	BeforeEach(func() {
		keypairGenerator = &FakeKeypairGenerator{}
		keypairUploader = &FakeKeypairUploader{}
		session = &FakeSession{}

		sessionProvider = &FakeSessionProvider{}
		sessionProvider.SessionCall.Returns.Session = session

		command = unsupported.NewCreateBoshAWSKeypair(keypairGenerator, keypairUploader, sessionProvider)
	})

	Describe("Execute", func() {
		It("generates a new keypair", func() {
			err := command.Execute(commands.GlobalFlags{})
			Expect(err).NotTo(HaveOccurred())
			Expect(keypairGenerator.GenerateCall.CallCount).To(Equal(1))
		})

		It("initializes a new session with the correct config", func() {
			err := command.Execute(commands.GlobalFlags{
				AWSAccessKeyID:     "some-aws-access-key-id",
				AWSSecretAccessKey: "some-aws-secret-access-key",
				AWSRegion:          "some-aws-region",
				EndpointOverride:   "some-endpoint-override",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(sessionProvider.SessionCall.Receives.Config).To(Equal(ec2.Config{
				AccessKeyID:      "some-aws-access-key-id",
				SecretAccessKey:  "some-aws-secret-access-key",
				Region:           "some-aws-region",
				EndpointOverride: "some-endpoint-override",
			}))
		})

		It("uploads the generated keypair", func() {
			keypairGenerator.GenerateCall.Returns.Keypair = ec2.Keypair{
				Name: "some-name",
				Key:  []byte("some-key"),
			}

			err := command.Execute(commands.GlobalFlags{})
			Expect(err).NotTo(HaveOccurred())
			Expect(keypairUploader.UploadCall.Receives.Session).To(Equal(session))
			Expect(keypairUploader.UploadCall.Receives.Keypair).To(Equal(ec2.Keypair{
				Name: "some-name",
				Key:  []byte("some-key"),
			}))
		})

		Context("failure cases", func() {
			It("returns an error when key generation fails", func() {
				keypairGenerator.GenerateCall.Returns.Error = errors.New("generate keys failed")
				err := command.Execute(commands.GlobalFlags{})

				Expect(err).To(MatchError("generate keys failed"))
			})

			It("returns an error when key upload fails", func() {
				keypairUploader.UploadCall.Returns.Error = errors.New("upload keys failed")
				err := command.Execute(commands.GlobalFlags{})

				Expect(err).To(MatchError("upload keys failed"))
			})
		})
	})
})
