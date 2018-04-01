package elbv2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevieve/leftovers/aws/elbv2"
	"github.com/genevieve/leftovers/aws/elbv2/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TargetGroup", func() {
	var (
		targetGroup elbv2.TargetGroup
		client      *fakes.TargetGroupsClient
		name        *string
		arn         *string
	)

	BeforeEach(func() {
		client = &fakes.TargetGroupsClient{}
		name = aws.String("the-name")
		arn = aws.String("the-arn")

		targetGroup = elbv2.NewTargetGroup(client, name, arn)
	})

	Describe("Delete", func() {
		It("deletes the target group", func() {
			err := targetGroup.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteTargetGroupCall.CallCount).To(Equal(1))
			Expect(client.DeleteTargetGroupCall.Receives.Input.TargetGroupArn).To(Equal(arn))
		})

		Context("when the client fails", func() {
			BeforeEach(func() {
				client.DeleteTargetGroupCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := targetGroup.Delete()
				Expect(err).To(MatchError("Delete: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(targetGroup.Name()).To(Equal("the-name"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(targetGroup.Type()).To(Equal("ELBV2 Target Group"))
		})
	})
})
