package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("gcp up", func() {
	var (
		stateStore *fakes.StateStore
		gcpUp      commands.GCPUp
	)

	BeforeEach(func() {
		stateStore = &fakes.StateStore{}
		gcpUp = commands.NewGCPUp(stateStore)
	})

	Context("Execute", func() {
		It("saves iaas gcp to the state", func() {
			err := gcpUp.Execute(storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(stateStore.SetCall.Receives.State).To(Equal(storage.State{
				IAAS: "gcp",
			}))
		})

		Context("failure cases", func() {
			It("returns an error when state store fails", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{Error: errors.New("set call failed")}}
				err := gcpUp.Execute(storage.State{})
				Expect(err).To(MatchError("set call failed"))
			})
		})
	})
})
