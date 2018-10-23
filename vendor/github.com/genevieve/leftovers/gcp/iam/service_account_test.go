package iam_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/iam"
	"github.com/genevieve/leftovers/gcp/iam/fakes"
	gcpcrm "google.golang.org/api/cloudresourcemanager/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceAccount", func() {
	var (
		client *fakes.ServiceAccountsClient
		logger *fakes.Logger
		name   string
		email  string

		serviceAccount iam.ServiceAccount
	)

	BeforeEach(func() {
		client = &fakes.ServiceAccountsClient{}
		logger = &fakes.Logger{}
		name = "banana"
		email = "banana@example.com"

		client.GetProjectIamPolicyCall.Returns.Output = &gcpcrm.Policy{
			Bindings: []*gcpcrm.Binding{},
		}

		serviceAccount = iam.NewServiceAccount(client, logger, name, email)
	})

	Describe("Delete", func() {
		It("deletes the service account", func() {
			err := serviceAccount.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.GetProjectIamPolicyCall.CallCount).To(Equal(1))

			Expect(client.DeleteServiceAccountCall.CallCount).To(Equal(1))
			Expect(client.DeleteServiceAccountCall.Receives.ServiceAccount).To(Equal(name))
		})

		Context("when there are bindings for the service account", func() {
			var updatedPolicy *gcpcrm.Policy

			BeforeEach(func() {
				client.GetProjectIamPolicyCall.Returns.Output = &gcpcrm.Policy{
					Bindings: []*gcpcrm.Binding{{
						Members: []string{"serviceAccount:other", "serviceAccount:banana@example.com"},
						Role:    "roles/some-role",
					}},
				}
				updatedPolicy = &gcpcrm.Policy{
					Bindings: []*gcpcrm.Binding{{
						Members: []string{"serviceAccount:other"},
						Role:    "roles/some-role",
					}},
				}
			})

			It("modifies the project policy to remove them and set the new policy", func() {
				err := serviceAccount.Delete()
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PrintfCall.Receives.Message).To(Equal("gcloud iam service-accounts remove-iam-policy-binding %s --member %s --role %s\n"))
				Expect(logger.PrintfCall.Receives.Arguments[0]).To(Equal("banana@example.com"))
				Expect(logger.PrintfCall.Receives.Arguments[1]).To(Equal("serviceAccount:banana@example.com"))
				Expect(logger.PrintfCall.Receives.Arguments[2]).To(Equal("roles/some-role"))

				Expect(client.SetProjectIamPolicyCall.CallCount).To(Equal(1))
				Expect(client.SetProjectIamPolicyCall.Receives.Input).To(Equal(updatedPolicy))
			})

			Context("when there are no more members in a binding", func() {
				BeforeEach(func() {
					client.GetProjectIamPolicyCall.Returns.Output = &gcpcrm.Policy{
						Bindings: []*gcpcrm.Binding{{
							Members: []string{"serviceAccount:banana@example.com"},
							Role:    "roles/some-role",
						}},
					}
				})

				It("removes the binding altogether", func() {
					err := serviceAccount.Delete()
					Expect(err).NotTo(HaveOccurred())

					Expect(client.SetProjectIamPolicyCall.CallCount).To(Equal(1))
					Expect(client.SetProjectIamPolicyCall.Receives.Input.Bindings).To(BeEmpty())
				})
			})

			Context("when the service account has more than one binding", func() {
				BeforeEach(func() {
					client.GetProjectIamPolicyCall.Returns.Output = &gcpcrm.Policy{
						Bindings: []*gcpcrm.Binding{
							{
								Members: []string{"serviceAccount:banana@example.com"},
								Role:    "roles/some-role",
							},
							{
								Members: []string{"user:apple", "serviceAccount:banana@example.com"},
								Role:    "roles/some-other-role",
							},
						},
					}
					updatedPolicy = &gcpcrm.Policy{
						Bindings: []*gcpcrm.Binding{
							{
								Members: []string{"user:apple"},
								Role:    "roles/some-other-role",
							},
						},
					}
				})

				It("removes the member from every binding", func() {
					err := serviceAccount.Delete()
					Expect(err).NotTo(HaveOccurred())

					Expect(client.SetProjectIamPolicyCall.CallCount).To(Equal(1))
					Expect(client.SetProjectIamPolicyCall.Receives.Input).To(Equal(updatedPolicy))
				})
			})
		})

		Context("when the client fails to get the project iam policy", func() {
			BeforeEach(func() {
				client.GetProjectIamPolicyCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := serviceAccount.Delete()
				Expect(err).To(MatchError("Remove IAM Policy Bindings: Get Project IAM Policy: the-error"))
			})
		})

		Context("when the client fails to set the project iam policy", func() {
			BeforeEach(func() {
				client.SetProjectIamPolicyCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := serviceAccount.Delete()
				Expect(err).To(MatchError("Remove IAM Policy Bindings: Set Project IAM Policy: the-error"))
			})
		})

		Context("when the client fails to delete the service account", func() {
			BeforeEach(func() {
				client.DeleteServiceAccountCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := serviceAccount.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(serviceAccount.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(serviceAccount.Type()).To(Equal("IAM Service Account"))
		})
	})
})
