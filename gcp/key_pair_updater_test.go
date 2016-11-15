package gcp_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"strings"

	compute "google.golang.org/api/compute/v1"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/gcp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
)

var _ = Describe("KeyPairUpdater", func() {
	var (
		keyPairUpdater    gcp.KeyPairUpdater
		gcpClientProvider *fakes.GCPClientProvider
		gcpClient         *fakes.GCPClient
		logger            *fakes.Logger
	)

	const (
		projectID = "some-project-id"
	)

	BeforeEach(func() {
		gcpClientProvider = &fakes.GCPClientProvider{}
		gcpClient = &fakes.GCPClient{}
		logger = &fakes.Logger{}

		gcpClientProvider.ClientCall.Returns.Client = gcpClient

		gcpClient.GetProjectCall.Returns.Project = &compute.Project{
			CommonInstanceMetadata: &compute.Metadata{
				Items: []*compute.MetadataItems{},
			},
		}
		keyPairUpdater = gcp.NewKeyPairUpdater(rand.Reader, rsa.GenerateKey, ssh.NewPublicKey, gcpClientProvider, logger)
	})

	It("generates a keypair", func() {
		keyPair, err := keyPairUpdater.Update(projectID)
		Expect(err).NotTo(HaveOccurred())
		Expect(keyPair.PrivateKey).NotTo(BeEmpty())
		Expect(keyPair.PublicKey).NotTo(BeEmpty())
		Expect(keyPair.PublicKey).NotTo(ContainSubstring("\n"))

		pemBlock, rest := pem.Decode([]byte(keyPair.PrivateKey))
		Expect(rest).To(HaveLen(0))
		Expect(pemBlock.Type).To(Equal("RSA PRIVATE KEY"))

		parsedPrivateKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
		Expect(err).NotTo(HaveOccurred())

		err = parsedPrivateKey.Validate()
		Expect(err).NotTo(HaveOccurred())

		newPublicKey, err := ssh.NewPublicKey(parsedPrivateKey.Public())
		Expect(err).NotTo(HaveOccurred())

		rawPublicKey := strings.TrimSuffix(string(ssh.MarshalAuthorizedKey(newPublicKey)), "\n")

		Expect(rawPublicKey).To(Equal(keyPair.PublicKey))
	})

	It("retrieves the project for the given project id", func() {
		_, err := keyPairUpdater.Update(projectID)
		Expect(err).NotTo(HaveOccurred())

		Expect(gcpClientProvider.ClientCall.CallCount).To(Equal(1))

		Expect(gcpClient.GetProjectCall.CallCount).To(Equal(1))
		Expect(gcpClient.GetProjectCall.Receives.ProjectID).To(Equal("some-project-id"))
	})

	It("updates common metadata for given project id", func() {
		_, err := keyPairUpdater.Update(projectID)
		Expect(err).NotTo(HaveOccurred())

		Expect(gcpClient.SetCommonInstanceMetadataCall.CallCount).To(Equal(1))
		Expect(gcpClient.SetCommonInstanceMetadataCall.Receives.ProjectID).To(Equal("some-project-id"))

		Expect(gcpClient.SetCommonInstanceMetadataCall.Receives.Metadata.Items).To(HaveLen(1))
		Expect(gcpClient.SetCommonInstanceMetadataCall.Receives.Metadata.Items[0].Key).To(Equal("sshKeys"))
		Expect(*gcpClient.SetCommonInstanceMetadataCall.Receives.Metadata.Items[0].Value).To(MatchRegexp(`vcap:ssh-rsa .* vcap\n$`))

		Expect(logger.StepCall.CallCount).To(Equal(1))
		Expect(logger.StepCall.Receives.Message).To(Equal(`Creating new ssh-keys for the project %q`))
		Expect(logger.StepCall.Receives.Arguments[0]).To(Equal("some-project-id"))
	})

	It("appends to the list of ssh-keys", func() {
		existingSSHKey := "my-user:ssh-rsa MY-PUBLIC-KEY my-user"
		someOtherValue := "some-other-value"
		gcpClient.GetProjectCall.Returns.Project.CommonInstanceMetadata = &compute.Metadata{
			Items: []*compute.MetadataItems{
				{
					Key:   "sshKeys",
					Value: &existingSSHKey,
				},
				{
					Key:   "some-other-key",
					Value: &someOtherValue,
				},
			},
		}
		_, err := keyPairUpdater.Update(projectID)
		Expect(err).NotTo(HaveOccurred())

		Expect(gcpClient.SetCommonInstanceMetadataCall.Receives.Metadata.Items).To(HaveLen(2))
		Expect(gcpClient.SetCommonInstanceMetadataCall.Receives.Metadata.Items[0].Key).To(Equal("sshKeys"))
		Expect(*gcpClient.SetCommonInstanceMetadataCall.Receives.Metadata.Items[0].Value).To(MatchRegexp(`my-user:ssh-rsa MY-PUBLIC-KEY my-user\nvcap:ssh-rsa .* vcap\n$`))

		Expect(logger.StepCall.CallCount).To(Equal(1))
		Expect(logger.StepCall.Receives.Message).To(Equal(`Appending new ssh-keys for the project %q`))
		Expect(logger.StepCall.Receives.Arguments[0]).To(Equal("some-project-id"))
	})

	Context("failure cases", func() {
		It("returns an error when the rsaKeyGenerator fails", func() {
			keyPairUpdater = gcp.NewKeyPairUpdater(rand.Reader,
				func(_ io.Reader, _ int) (*rsa.PrivateKey, error) {
					return nil, errors.New("rsa key generator failed")
				},
				ssh.NewPublicKey, gcpClientProvider, logger)

			_, err := keyPairUpdater.Update(projectID)
			Expect(err).To(MatchError("rsa key generator failed"))
		})

		It("returns an error when the ssh public key generator fails", func() {
			keyPairUpdater = gcp.NewKeyPairUpdater(rand.Reader, rsa.GenerateKey,
				func(_ interface{}) (ssh.PublicKey, error) {
					return nil, errors.New("ssh public key gen failed")
				}, gcpClientProvider, logger)

			_, err := keyPairUpdater.Update(projectID)
			Expect(err).To(MatchError("ssh public key gen failed"))
		})

		It("returns an error when project could not be found", func() {
			gcpClient.GetProjectCall.Returns.Error = errors.New("project could not be found")

			_, err := keyPairUpdater.Update(projectID)
			Expect(err).To(MatchError("project could not be found"))
		})

		It("returns an error when set common instance metadata fails", func() {
			gcpClient.SetCommonInstanceMetadataCall.Returns.Error = errors.New("updating ssh-key failed")

			_, err := keyPairUpdater.Update(projectID)
			Expect(err).To(MatchError("updating ssh-key failed"))
		})
	})
})
