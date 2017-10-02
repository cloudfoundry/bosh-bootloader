package commands_test

import (
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AzureUp", func() {
	var (
		azureUp       commands.AzureUp
		incomingState storage.State
	)

	BeforeEach(func() {
		incomingState = storage.State{
			Azure: storage.Azure{
				SubscriptionID: "subscription-id",
			},
		}
		azureUp = commands.NewAzureUp()
	})

	Describe("Execute", func() {
		It("returns the state", func() {
			returnedState, err := azureUp.Execute(incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(returnedState).To(Equal(incomingState))
		})
	})
})
