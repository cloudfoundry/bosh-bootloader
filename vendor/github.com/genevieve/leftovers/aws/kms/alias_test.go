package kms_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevieve/leftovers/aws/kms"
	"github.com/genevieve/leftovers/aws/kms/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Alias", func() {
	var (
		alias  kms.Alias
		client *fakes.AliasesClient
		name   *string
	)

	BeforeEach(func() {
		client = &fakes.AliasesClient{}
		name = aws.String("the-name")

		alias = kms.NewAlias(client, name)
	})

	Describe("Delete", func() {
		It("deletes the alias", func() {
			err := alias.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteAliasCall.CallCount).To(Equal(1))
			Expect(client.DeleteAliasCall.Receives.Input.AliasName).To(Equal(name))
		})

		Context("when the client fails", func() {
			BeforeEach(func() {
				client.DeleteAliasCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := alias.Delete()
				Expect(err).To(MatchError("Delete: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(alias.Name()).To(Equal("the-name"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(alias.Type()).To(Equal("KMS Alias"))
		})
	})
})
