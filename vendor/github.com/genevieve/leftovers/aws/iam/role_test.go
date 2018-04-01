package iam_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevieve/leftovers/aws/iam"
	"github.com/genevieve/leftovers/aws/iam/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Role", func() {
	var (
		role     iam.Role
		client   *fakes.RolesClient
		policies *fakes.RolePolicies
		name     *string
	)

	BeforeEach(func() {
		client = &fakes.RolesClient{}
		policies = &fakes.RolePolicies{}
		name = aws.String("the-name")

		role = iam.NewRole(client, policies, name)
	})

	Describe("Delete", func() {
		It("deletes the role", func() {
			err := role.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(policies.DeleteCall.CallCount).To(Equal(1))
			Expect(policies.DeleteCall.Receives.RoleName).To(Equal(*name))

			Expect(client.DeleteRoleCall.CallCount).To(Equal(1))
			Expect(client.DeleteRoleCall.Receives.Input.RoleName).To(Equal(name))
		})

		Context("when deleting the role's policies fails", func() {
			BeforeEach(func() {
				policies.DeleteCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := role.Delete()
				Expect(err).To(MatchError("Delete policies: banana"))
			})
		})

		Context("when the client fails", func() {
			BeforeEach(func() {
				client.DeleteRoleCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := role.Delete()
				Expect(err).To(MatchError("Delete: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(role.Name()).To(Equal("the-name"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(role.Type()).To(Equal("IAM Role"))
		})
	})
})
