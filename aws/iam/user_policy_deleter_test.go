package iam_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserPolicyDeleter", func() {
	var (
		iamClient *fakes.Client
		deleter   iam.UserPolicyDeleter
	)

	BeforeEach(func() {
		iamClient = &fakes.Client{}
		deleter = iam.NewUserPolicyDeleter(iamClient)
	})

	Describe("Delete", func() {
		It("deletes the user policy with the given username and policy name", func() {
			iamClient.DeleteUserPolicyReturns(&awsiam.DeleteUserPolicyOutput{}, nil)

			err := deleter.Delete("some-username", "some-policy-name")
			Expect(err).NotTo(HaveOccurred())

			input := iamClient.DeleteUserPolicyArgsForCall(0)

			Expect(input.UserName).To(Equal(aws.String("some-username")))
			Expect(input.PolicyName).To(Equal(aws.String("some-policy-name")))
		})

		Context("failure cases", func() {
			It("returns an error when it fails to delete", func() {
				iamClient.DeleteUserPolicyReturns(nil, errors.New("failed to delete user policy"))

				err := deleter.Delete("some-username", "some-policy-name")
				Expect(err).To(MatchError("failed to delete user policy"))
			})
		})
	})
})
