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

var _ = Describe("Instance", func() {
	var (
		instance     ec2.Instance
		client       *fakes.InstancesClient
		resourceTags *fakes.ResourceTags
		id           *string
		keyName      *string
	)

	BeforeEach(func() {
		client = &fakes.InstancesClient{}
		resourceTags = &fakes.ResourceTags{}
		id = aws.String("the-id")
		keyName = aws.String("the-key-name")
		tags := []*awsec2.Tag{}

		instance = ec2.NewInstance(client, resourceTags, id, keyName, tags)
	})

	Describe("Delete", func() {
		It("terminates the instance", func() {
			err := instance.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.TerminateInstancesCall.CallCount).To(Equal(1))
			Expect(client.TerminateInstancesCall.Receives.Input.InstanceIds).To(HaveLen(1))
			Expect(client.TerminateInstancesCall.Receives.Input.InstanceIds[0]).To(Equal(id))

			Expect(resourceTags.DeleteCall.CallCount).To(Equal(1))
			Expect(resourceTags.DeleteCall.Receives.ResourceType).To(Equal("instance"))
			Expect(resourceTags.DeleteCall.Receives.ResourceId).To(Equal("the-id"))
		})

		Context("when the client fails", func() {
			BeforeEach(func() {
				client.TerminateInstancesCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := instance.Delete()
				Expect(err).To(MatchError("Terminate: banana"))
			})
		})

		Context("when resource tags fails", func() {
			BeforeEach(func() {
				resourceTags.DeleteCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := instance.Delete()
				Expect(err).To(MatchError("Delete resource tags: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(instance.Name()).To(Equal("the-id (KeyPairName:the-key-name)"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(instance.Type()).To(Equal("EC2 Instance"))
		})
	})
})
