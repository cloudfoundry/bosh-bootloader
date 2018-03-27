package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/ec2/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SecurityGroup", func() {
	var (
		securityGroup ec2.SecurityGroup
		client        *fakes.SecurityGroupsClient
		id            *string
		groupName     *string
	)

	BeforeEach(func() {
		client = &fakes.SecurityGroupsClient{}
		id = aws.String("the-id")
		groupName = aws.String("the-group-name")
		tags := []*awsec2.Tag{}

		securityGroup = ec2.NewSecurityGroup(client, id, groupName, tags)
	})

	Describe("Delete", func() {
		It("deletes the security group", func() {
			err := securityGroup.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteSecurityGroupCall.CallCount).To(Equal(1))
			Expect(client.DeleteSecurityGroupCall.Receives.Input.GroupId).To(Equal(id))
		})

		Context("the client fails", func() {
			BeforeEach(func() {
				client.DeleteSecurityGroupCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := securityGroup.Delete()
				Expect(err).To(MatchError("FAILED deleting security group the-group-name: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(securityGroup.Name()).To(Equal("the-group-name"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(securityGroup.Type()).To(Equal("security group"))
		})
	})
})
