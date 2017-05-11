package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rotate", func() {
	var (
		stateStore     *fakes.StateStore
		keyPairManager *fakes.KeyPairManager
		command        commands.Rotate
		incomingState  storage.State
	)

	Describe("Execute", func() {
		BeforeEach(func() {
			stateStore = &fakes.StateStore{}
			keyPairManager = &fakes.KeyPairManager{}

			command = commands.NewRotate(stateStore, keyPairManager)
			incomingState = storage.State{
				KeyPair: storage.KeyPair{
					Name:       "some-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
			}

			keyPairManager.RotateCall.Returns.State = storage.State{
				KeyPair: storage.KeyPair{
					Name:       "some-new-name",
					PrivateKey: "some-new-private-key",
					PublicKey:  "some-new-public-key",
				},
			}
		})

		It("rotates the keys", func() {
			err := command.Execute([]string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(keyPairManager.RotateCall.CallCount).To(Equal(1))
			Expect(keyPairManager.RotateCall.Receives.State).To(Equal(incomingState))
			Expect(stateStore.SetCall.CallCount).To(Equal(1))
			Expect(stateStore.SetCall.Receives[0].State).To(Equal(storage.State{
				KeyPair: storage.KeyPair{
					Name:       "some-new-name",
					PrivateKey: "some-new-private-key",
					PublicKey:  "some-new-public-key",
				},
			}))
		})

		Context("failure cases", func() {
			It("returns an error when key pair manager rotate fails", func() {
				keyPairManager.RotateCall.Returns.Error = errors.New("failed to rotate")
				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("failed to rotate"))
			})

			It("returns an error when stateStore set fails", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to set")}}
				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("failed to set"))
			})
		})
	})
})
