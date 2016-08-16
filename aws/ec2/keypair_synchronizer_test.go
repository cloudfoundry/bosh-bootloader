package ec2_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KeyPairSynchronizer", func() {
	var (
		synchronizer   ec2.KeyPairSynchronizer
		keyPairManager *fakes.KeyPairManager
	)

	BeforeEach(func() {
		keyPairManager = &fakes.KeyPairManager{}
		keyPairManager.SyncCall.Returns.KeyPair = ec2.KeyPair{
			Name:       "updated-keypair-name",
			PrivateKey: "updated-private-key",
			PublicKey:  "updated-public-key",
		}

		synchronizer = ec2.NewKeyPairSynchronizer(keyPairManager)
	})

	It("syncs the keypair", func() {
		keyPair, err := synchronizer.Sync(ec2.KeyPair{
			Name:       "some-keypair-name",
			PrivateKey: "some-private-key",
			PublicKey:  "some-public-key",
		}, "some-env-id")
		Expect(err).NotTo(HaveOccurred())

		Expect(keyPairManager.SyncCall.Receives.KeyPair).To(Equal(ec2.KeyPair{
			Name:       "some-keypair-name",
			PrivateKey: "some-private-key",
			PublicKey:  "some-public-key",
		}))

		Expect(keyPairManager.SyncCall.Receives.EnvID).To(Equal("some-env-id"))

		Expect(keyPair).To(Equal(ec2.KeyPair{
			Name:       "updated-keypair-name",
			PublicKey:  "updated-public-key",
			PrivateKey: "updated-private-key",
		}))
	})

	Context("failure cases", func() {
		Context("when the key pair cannot by synced", func() {
			It("returns an error", func() {
				keyPairManager.SyncCall.Returns.Error = errors.New("failed to sync")

				_, err := synchronizer.Sync(ec2.KeyPair{}, "")
				Expect(err).To(MatchError("failed to sync"))
			})
		})
	})
})
