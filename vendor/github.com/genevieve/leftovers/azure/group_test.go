package azure_test

import (
	"errors"

	"github.com/genevieve/leftovers/azure"
	"github.com/genevieve/leftovers/azure/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Group", func() {
	var (
		client *fakes.GroupsClient
		name   string

		group azure.Group
	)

	BeforeEach(func() {
		client = &fakes.GroupsClient{}
		name = "banana-group"

		group = azure.NewGroup(client, &name)
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			errChan := make(chan error, 1)
			errChan <- nil
			client.DeleteCall.Returns.Error = errChan
		})

		It("deletes resource groups", func() {
			err := group.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteCall.CallCount).To(Equal(1))
			Expect(client.DeleteCall.Receives.Name).To(Equal("banana-group"))
		})

		Context("when client fails to delete the resource group", func() {
			BeforeEach(func() {
				errChan := make(chan error, 1)
				errChan <- errors.New("some error")
				client.DeleteCall.Returns.Error = errChan
			})

			It("logs the error", func() {
				err := group.Delete()
				Expect(err).To(MatchError("Delete: some error"))
			})
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(group.Type()).To(Equal("Resource Group"))
		})
	})
})
