package iam_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/genevieve/leftovers/aws/iam"
	"github.com/genevieve/leftovers/aws/iam/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstanceProfile", func() {
	var (
		instanceProfile iam.InstanceProfile
		client          *fakes.InstanceProfilesClient
		name            *string
	)

	BeforeEach(func() {
		client = &fakes.InstanceProfilesClient{}
		name = aws.String("the-name")
		roles := []*awsiam.Role{}

		instanceProfile = iam.NewInstanceProfile(client, name, roles)
	})

	Describe("Delete", func() {
		It("deletes the instance profile", func() {
			err := instanceProfile.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteInstanceProfileCall.CallCount).To(Equal(1))
			Expect(client.DeleteInstanceProfileCall.Receives.Input.InstanceProfileName).To(Equal(name))
		})

		Context("when the client fails", func() {
			BeforeEach(func() {
				client.DeleteInstanceProfileCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := instanceProfile.Delete()
				Expect(err).To(MatchError("FAILED deleting instance profile the-name: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(instanceProfile.Name()).To(Equal("the-name"))
		})
	})
})
