package iam_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/iam"
	"github.com/genevieve/leftovers/gcp/iam/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceAccount", func() {
	var (
		client *fakes.ServiceAccountsClient
		name   string

		instance iam.ServiceAccount
	)

	BeforeEach(func() {
		client = &fakes.ServiceAccountsClient{}
		name = "banana"

		instance = iam.NewServiceAccount(client, name)
	})

	Describe("Delete", func() {
		It("deletes the instance", func() {
			err := instance.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteServiceAccountCall.CallCount).To(Equal(1))
			Expect(client.DeleteServiceAccountCall.Receives.ServiceAccount).To(Equal(name))
		})

		Context("when the client fails to delete the instance", func() {
			BeforeEach(func() {
				client.DeleteServiceAccountCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := instance.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(instance.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(instance.Type()).To(Equal("IAM Service Account"))
		})
	})
})
