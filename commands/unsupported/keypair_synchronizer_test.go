package unsupported_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KeyPairSynchronizer", func() {
	var (
		synchronizer   unsupported.KeyPairSynchronizer
		keyPairManager *fakes.KeyPairManager
		ec2Client      *fakes.EC2Client
	)

	BeforeEach(func() {
		ec2Client = &fakes.EC2Client{}
		keyPairManager = &fakes.KeyPairManager{}
		keyPairManager.SyncCall.Returns.KeyPair = ec2.KeyPair{
			Name:       "updated-keypair-name",
			PrivateKey: []byte("updated-private-key"),
			PublicKey:  []byte("updated-public-key"),
		}

		synchronizer = unsupported.NewKeyPairSynchronizer(keyPairManager)
	})

	It("syncs the keypair", func() {
		keyPair, err := synchronizer.Sync(unsupported.KeyPair{
			Name:       "some-keypair-name",
			PrivateKey: "some-private-key",
			PublicKey:  "some-public-key",
		}, ec2Client)
		Expect(err).NotTo(HaveOccurred())

		Expect(keyPairManager.SyncCall.Receives.EC2Client).To(Equal(ec2Client))
		Expect(keyPairManager.SyncCall.Receives.KeyPair).To(Equal(ec2.KeyPair{
			Name:       "some-keypair-name",
			PrivateKey: []byte("some-private-key"),
			PublicKey:  []byte("some-public-key"),
		}))

		Expect(keyPair).To(Equal(unsupported.KeyPair{
			Name:       "updated-keypair-name",
			PublicKey:  "updated-public-key",
			PrivateKey: "updated-private-key",
		}))
	})

	Context("failure cases", func() {
		Context("when the key pair cannot by synced", func() {
			It("returns an error", func() {
				keyPairManager.SyncCall.Returns.Error = errors.New("failed to sync")

				_, err := synchronizer.Sync(unsupported.KeyPair{}, ec2Client)
				Expect(err).To(MatchError("failed to sync"))
			})
		})
	})
})
