package gcp_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"

	compute "google.golang.org/api/compute/v1"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/gcp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
)

var _ = Describe("KeyPairUpdater", func() {
	var (
		keyPairUpdater     gcp.KeyPairUpdater
		gcpProvider        *fakes.GCPProvider
		gcpService         *fakes.GCPService
		gcpProjectsService *fakes.GCPProjectsService
		project            compute.Project
		logger             *fakes.Logger
	)

	const (
		projectID = "some-project-id"
	)

	BeforeEach(func() {
		gcpProvider = &fakes.GCPProvider{}
		gcpService = &fakes.GCPService{}
		gcpProjectsService = &fakes.GCPProjectsService{}
		logger = &fakes.Logger{}

		project.CommonInstanceMetadata = &compute.Metadata{
			Items: []*compute.MetadataItems{},
		}

		gcpProjectsService.GetCall.Returns.Project = &project

		gcpService.GetProjectsServiceCall.Returns.ProjectsService = gcpProjectsService
		gcpProvider.GetServiceCall.Returns.Service = gcpService
		keyPairUpdater = gcp.NewKeyPairUpdater(rand.Reader, rsa.GenerateKey, ssh.NewPublicKey, gcpProvider, logger)
	})

	It("generates a keypair", func() {
		keyPair, err := keyPairUpdater.Update(projectID)
		Expect(err).NotTo(HaveOccurred())
		Expect(keyPair.PrivateKey).NotTo(BeEmpty())
		Expect(keyPair.PublicKey).NotTo(BeEmpty())

		pemBlock, rest := pem.Decode([]byte(keyPair.PrivateKey))
		Expect(rest).To(HaveLen(0))
		Expect(pemBlock.Type).To(Equal("RSA PRIVATE KEY"))

		parsedPrivateKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
		Expect(err).NotTo(HaveOccurred())

		err = parsedPrivateKey.Validate()
		Expect(err).NotTo(HaveOccurred())

		newPublicKey, err := ssh.NewPublicKey(parsedPrivateKey.Public())
		Expect(err).NotTo(HaveOccurred())

		Expect(string(ssh.MarshalAuthorizedKey(newPublicKey))).To(Equal(keyPair.PublicKey))
	})

	It("retrieves the project for the given project id", func() {
		_, err := keyPairUpdater.Update(projectID)
		Expect(err).NotTo(HaveOccurred())

		Expect(gcpProvider.GetServiceCall.CallCount).To(Equal(1))
		Expect(gcpService.GetProjectsServiceCall.CallCount).To(Equal(1))
		Expect(gcpProjectsService.GetCall.CallCount).To(Equal(1))
		Expect(gcpProjectsService.GetCall.Receives.ProjectID).To(Equal("some-project-id"))
	})

	It("updates ssh-key on gcp for given project id", func() {
		_, err := keyPairUpdater.Update(projectID)
		Expect(err).NotTo(HaveOccurred())

		Expect(gcpProjectsService.SetCommonInstanceMetadataCall.CallCount).To(Equal(1))
		Expect(gcpProjectsService.SetCommonInstanceMetadataCall.Receives.ProjectID).To(Equal("some-project-id"))
		Expect(gcpProjectsService.SetCommonInstanceMetadataCall.Receives.Metadata.Items).To(HaveLen(1))
		Expect(gcpProjectsService.SetCommonInstanceMetadataCall.Receives.Metadata.Items[0].Key).To(Equal("sshKeys"))
		Expect(*gcpProjectsService.SetCommonInstanceMetadataCall.Receives.Metadata.Items[0].Value).To(MatchRegexp("vcap:ssh-rsa .* vcap"))

		Expect(logger.StepCall.CallCount).To(Equal(1))
		Expect(logger.StepCall.Receives.Message).To(Equal(`Creating new ssh-keys for the project %q`))
		Expect(logger.StepCall.Receives.Arguments[0]).To(Equal("some-project-id"))
	})

	It("appends to the list of ssh-keys", func() {
		existingSSHKey := "my-user:ssh-rsa MY-PUBLIC-KEY my-user"
		someOtherValue := "some-other-value"
		project.CommonInstanceMetadata = &compute.Metadata{
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

		Expect(gcpProjectsService.SetCommonInstanceMetadataCall.Receives.Metadata.Items).To(HaveLen(2))
		Expect(gcpProjectsService.SetCommonInstanceMetadataCall.Receives.Metadata.Items[0].Key).To(Equal("sshKeys"))
		Expect(*gcpProjectsService.SetCommonInstanceMetadataCall.Receives.Metadata.Items[0].Value).To(MatchRegexp("my-user:ssh-rsa MY-PUBLIC-KEY my-user\nvcap:ssh-rsa .* vcap"))

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
				ssh.NewPublicKey, gcpProvider, logger)

			_, err := keyPairUpdater.Update(projectID)
			Expect(err).To(MatchError("rsa key generator failed"))
		})

		It("returns an error when the ssh public key generator fails", func() {
			keyPairUpdater = gcp.NewKeyPairUpdater(rand.Reader, rsa.GenerateKey,
				func(_ interface{}) (ssh.PublicKey, error) {
					return nil, errors.New("ssh public key gen failed")
				}, gcpProvider, logger)

			_, err := keyPairUpdater.Update(projectID)
			Expect(err).To(MatchError("ssh public key gen failed"))
		})

		It("returns an error when project could not be found", func() {
			gcpProjectsService.GetCall.Returns.Error = errors.New("project could not be found")

			_, err := keyPairUpdater.Update(projectID)
			Expect(err).To(MatchError("project could not be found"))
		})

		It("return an error when updating the ssh-key fails", func() {
			gcpProjectsService.SetCommonInstanceMetadataCall.Returns.Error = errors.New("updating ssh-key failed")

			_, err := keyPairUpdater.Update(projectID)
			Expect(err).To(MatchError("updating ssh-key failed"))
		})
	})
})
