package iam_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevievelesperance/leftovers/aws/iam"
	"github.com/genevievelesperance/leftovers/aws/iam/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Policy", func() {
	var (
		policy iam.Policy
		client *fakes.PoliciesClient
		name   *string
		arn    *string
	)

	BeforeEach(func() {
		client = &fakes.PoliciesClient{}
		name = aws.String("the-name")
		arn = aws.String("the-arn")

		policy = iam.NewPolicy(client, name, arn)
	})

	It("deletes the policy", func() {
		err := policy.Delete()
		Expect(err).NotTo(HaveOccurred())

		Expect(client.DeletePolicyCall.CallCount).To(Equal(1))
		Expect(client.DeletePolicyCall.Receives.Input.PolicyArn).To(Equal(arn))
	})

	Context("when the client fails", func() {
		BeforeEach(func() {
			client.DeletePolicyCall.Returns.Error = errors.New("banana")
		})

		It("returns the error", func() {
			err := policy.Delete()
			Expect(err).To(MatchError("FAILED deleting policy the-name: banana"))
		})
	})
})
