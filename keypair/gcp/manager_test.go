package gcp_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/keypair/gcp"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {
	var (
		keyPairUpdater *fakes.GCPKeyPairUpdater
		keyPairManager gcp.Manager
	)

	BeforeEach(func() {
		keyPairUpdater = &fakes.GCPKeyPairUpdater{}
		keyPairUpdater.UpdateCall.Returns.KeyPair = storage.KeyPair{
			PrivateKey: "some-private-key",
			PublicKey:  "some-public-key",
		}

		keyPairManager = gcp.NewManager(keyPairUpdater)
	})

	Context("when keypair is empty", func() {
		It("calls keypair updater and saves the keypair to the state", func() {
			state, err := keyPairManager.Sync(storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(keyPairUpdater.UpdateCall.CallCount).To(Equal(1))
			Expect(state).To(Equal(storage.State{
				KeyPair: storage.KeyPair{
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
			}))
		})

		Context("failure cases", func() {
			It("returns an error when key pair updater fails", func() {
				keyPairUpdater.UpdateCall.Returns.Error = errors.New("failed to update")
				_, err := keyPairManager.Sync(storage.State{})
				Expect(err).To(MatchError("failed to update"))
			})
		})
	})

	Context("when keypair exists", func() {
		It("no-ops and returns provided state", func() {
			state, err := keyPairManager.Sync(storage.State{
				KeyPair: storage.KeyPair{
					PrivateKey: "some-existing-private-key",
					PublicKey:  "some-existing-public-key",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(keyPairUpdater.UpdateCall.CallCount).To(Equal(0))
			Expect(state).To(Equal(storage.State{
				KeyPair: storage.KeyPair{
					PrivateKey: "some-existing-private-key",
					PublicKey:  "some-existing-public-key",
				},
			}))
		})
	})
})
