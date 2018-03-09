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
		ingressPerms := []*awsec2.IpPermission{}
		egressPerms := []*awsec2.IpPermission{}

		securityGroup = ec2.NewSecurityGroup(client, id, groupName, tags, ingressPerms, egressPerms)
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

		Context("when the client has ingress rules", func() {
			BeforeEach(func() {
				ingressPerms := []*awsec2.IpPermission{{
					IpProtocol: aws.String("tcp"),
				}}

				securityGroup = ec2.NewSecurityGroup(client, id, groupName, []*awsec2.Tag{}, ingressPerms, []*awsec2.IpPermission{})
			})

			It("revokes them", func() {
				err := securityGroup.Delete()
				Expect(err).NotTo(HaveOccurred())

				Expect(client.RevokeSecurityGroupIngressCall.CallCount).To(Equal(1))
				Expect(client.RevokeSecurityGroupIngressCall.Receives.Input.GroupId).To(Equal(aws.String("the-group-name")))
				Expect(client.RevokeSecurityGroupIngressCall.Receives.Input.IpPermissions[0].IpProtocol).To(Equal(aws.String("tcp")))
			})

			Context("when the client fails to revoke ingress rules", func() {
				BeforeEach(func() {
					client.RevokeSecurityGroupIngressCall.Returns.Error = errors.New("some error")
				})

				It("logs the error", func() {
					err := securityGroup.Delete()
					Expect(err).To(MatchError("ERROR revoking security group ingress for the-group-name: some error\n"))
				})
			})
		})

		Context("when the client has egress rules", func() {
			BeforeEach(func() {
				egressPerms := []*awsec2.IpPermission{{
					IpProtocol: aws.String("tcp"),
				}}
				securityGroup = ec2.NewSecurityGroup(client, id, groupName, []*awsec2.Tag{}, []*awsec2.IpPermission{}, egressPerms)
			})

			It("revokes them", func() {
				err := securityGroup.Delete()
				Expect(err).NotTo(HaveOccurred())

				Expect(client.RevokeSecurityGroupEgressCall.CallCount).To(Equal(1))
				Expect(client.RevokeSecurityGroupEgressCall.Receives.Input.GroupId).To(Equal(aws.String("the-group-name")))
				Expect(client.RevokeSecurityGroupEgressCall.Receives.Input.IpPermissions[0].IpProtocol).To(Equal(aws.String("tcp")))
			})

			Context("when the client fails to revoke egress rules", func() {
				BeforeEach(func() {
					client.RevokeSecurityGroupEgressCall.Returns.Error = errors.New("some error")
				})

				It("logs the error", func() {
					err := securityGroup.Delete()
					Expect(err).To(MatchError("ERROR revoking security group egress for the-group-name: some error\n"))
				})
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(securityGroup.Name()).To(Equal("the-group-name"))
		})
	})

	Describe("Type", func() {
		It("returns \"security group\"", func() {
			Expect(securityGroup.Type()).To(Equal("security group"))
		})
	})
})
