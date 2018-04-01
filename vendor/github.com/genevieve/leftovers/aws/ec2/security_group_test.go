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
		client       *fakes.SecurityGroupsClient
		logger       *fakes.Logger
		resourceTags *fakes.ResourceTags
		id           *string
		groupName    *string
		tags         []*awsec2.Tag
		ingress      []*awsec2.IpPermission
		egress       []*awsec2.IpPermission

		securityGroup ec2.SecurityGroup
	)

	BeforeEach(func() {
		client = &fakes.SecurityGroupsClient{}
		logger = &fakes.Logger{}
		resourceTags = &fakes.ResourceTags{}
		id = aws.String("the-id")
		groupName = aws.String("the-group-name")
		tags = []*awsec2.Tag{}
		ingress = []*awsec2.IpPermission{}
		egress = []*awsec2.IpPermission{}

		securityGroup = ec2.NewSecurityGroup(client, logger, resourceTags, id, groupName, tags, ingress, egress)
	})

	Describe("Delete", func() {
		It("deletes the security group", func() {
			err := securityGroup.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteSecurityGroupCall.CallCount).To(Equal(1))
			Expect(client.DeleteSecurityGroupCall.Receives.Input.GroupId).To(Equal(id))
		})

		Context("when the client fails", func() {
			BeforeEach(func() {
				client.DeleteSecurityGroupCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := securityGroup.Delete()
				Expect(err).To(MatchError("Delete: banana"))
			})
		})

		Context("when the security group has ingress rules", func() {
			BeforeEach(func() {
				ingress = []*awsec2.IpPermission{{IpProtocol: aws.String("tcp")}}
				securityGroup = ec2.NewSecurityGroup(client, logger, resourceTags, id, groupName, tags, ingress, egress)
			})

			It("revokes them", func() {
				err := securityGroup.Delete()
				Expect(err).NotTo(HaveOccurred())

				Expect(client.RevokeSecurityGroupIngressCall.CallCount).To(Equal(1))
				Expect(client.RevokeSecurityGroupIngressCall.Receives.Input.GroupId).To(Equal(aws.String("the-id")))
				Expect(client.RevokeSecurityGroupIngressCall.Receives.Input.IpPermissions[0].IpProtocol).To(Equal(aws.String("tcp")))
			})

			Context("when the client fails to revoke ingress rules", func() {
				BeforeEach(func() {
					client.RevokeSecurityGroupIngressCall.Returns.Error = errors.New("some error")
				})

				It("returns the error", func() {
					err := securityGroup.Delete()
					Expect(err).To(MatchError("Revoke ingress: some error"))
				})
			})
		})

		Context("when the security group has egress rules", func() {
			BeforeEach(func() {
				egress = []*awsec2.IpPermission{{IpProtocol: aws.String("tcp")}}
				securityGroup = ec2.NewSecurityGroup(client, logger, resourceTags, id, groupName, tags, ingress, egress)
			})

			It("revokes them", func() {
				err := securityGroup.Delete()
				Expect(err).NotTo(HaveOccurred())

				Expect(client.RevokeSecurityGroupEgressCall.CallCount).To(Equal(1))
				Expect(client.RevokeSecurityGroupEgressCall.Receives.Input.GroupId).To(Equal(aws.String("the-id")))
				Expect(client.RevokeSecurityGroupEgressCall.Receives.Input.IpPermissions[0].IpProtocol).To(Equal(aws.String("tcp")))
			})

			Context("when the client fails to revoke egress rules", func() {
				BeforeEach(func() {
					client.RevokeSecurityGroupEgressCall.Returns.Error = errors.New("some error")
				})

				It("returns the error", func() {
					err := securityGroup.Delete()
					Expect(err).To(MatchError("Revoke egress: some error"))
				})
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(securityGroup.Name()).To(Equal("the-group-name"))
		})

		Context("when the security group has tags", func() {
			BeforeEach(func() {
				tags = []*awsec2.Tag{{Key: aws.String("the-key"), Value: aws.String("the-value")}}
				securityGroup = ec2.NewSecurityGroup(client, logger, resourceTags, id, groupName, tags, ingress, egress)
			})
			It("uses the tag in the name", func() {
				Expect(securityGroup.Name()).To(Equal("the-group-name (the-key:the-value)"))
			})
		})

	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(securityGroup.Type()).To(Equal("EC2 Security Group"))
		})
	})
})
