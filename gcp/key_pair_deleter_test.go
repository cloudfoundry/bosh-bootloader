package gcp_test

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/gcp"

	compute "google.golang.org/api/compute/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KeyPairDeleter", func() {
	var (
		deleter           gcp.KeyPairDeleter
		client            *fakes.GCPClient
		gcpClientProvider *fakes.GCPClientProvider
		logger            *fakes.Logger
	)

	BeforeEach(func() {
		gcpClientProvider = &fakes.GCPClientProvider{}
		client = &fakes.GCPClient{}
		logger = &fakes.Logger{}
		gcpClientProvider.ClientCall.Returns.Client = client
		deleter = gcp.NewKeyPairDeleter(gcpClientProvider, logger)
	})

	It("deletes the gcp keypair", func() {
		publicKey := "ssh-rsa some-public-key"
		sshKeysValue := fmt.Sprintf("vcap:%s vcap\nsomeuser:ssh-rsa some-other-public-key someuser", publicKey)
		client.GetProjectCall.Returns.Project = &compute.Project{
			CommonInstanceMetadata: &compute.Metadata{
				Items: []*compute.MetadataItems{
					{
						Key:   "sshKeys",
						Value: &sshKeysValue,
					},
				},
			},
		}
		err := deleter.Delete("some-project-id", publicKey)
		Expect(err).NotTo(HaveOccurred())

		Expect(gcpClientProvider.ClientCall.CallCount).To(Equal(1))

		Expect(client.GetProjectCall.Receives.ProjectID).To(Equal("some-project-id"))

		Expect(client.SetCommonInstanceMetadataCall.Receives.ProjectID).To(Equal("some-project-id"))

		expectedSSHKeysValue := "someuser:ssh-rsa some-other-public-key someuser"
		Expect(*client.SetCommonInstanceMetadataCall.Receives.Metadata).To(Equal(compute.Metadata{
			Items: []*compute.MetadataItems{
				{
					Key:   "sshKeys",
					Value: &expectedSSHKeysValue,
				},
			},
		}))

		Expect(logger.StepCall.Receives.Message).To(Equal("deleting keypair"))
	})

	Context("when keypair does not exist in project metadata", func() {
		It("does not set common instance metadata", func() {
			sshKeysValue := "vcap:ssh-rsa something vcap\n"
			client.GetProjectCall.Returns.Project = &compute.Project{
				CommonInstanceMetadata: &compute.Metadata{
					Items: []*compute.MetadataItems{
						{
							Key:   "sshKeys",
							Value: &sshKeysValue,
						},
					},
				},
			}

			err := deleter.Delete("some-project-id", "non-existent-pub-key")
			Expect(err).NotTo(HaveOccurred())

			Expect(client.SetCommonInstanceMetadataCall.CallCount).To(Equal(0))
		})
	})

	Context("failure cases", func() {
		It("returns an error when the project cannot be retrieved", func() {
			client.GetProjectCall.Returns.Error = errors.New("project retrieval failed")

			err := deleter.Delete("", "")
			Expect(err).To(MatchError("project retrieval failed"))
		})

		It("returns an error when setting common instance metadata fails", func() {
			sshKeysValue := fmt.Sprintf("vcap:%s vcap", "ssh-rsa some-pub-key")
			client.GetProjectCall.Returns.Project = &compute.Project{
				CommonInstanceMetadata: &compute.Metadata{
					Items: []*compute.MetadataItems{
						{
							Key:   "sshKeys",
							Value: &sshKeysValue,
						},
					},
				},
			}
			client.SetCommonInstanceMetadataCall.Returns.Error = errors.New("set common metadata failed")

			err := deleter.Delete("", "ssh-rsa some-pub-key")
			Expect(err).To(MatchError("set common metadata failed"))
		})
	})
})
