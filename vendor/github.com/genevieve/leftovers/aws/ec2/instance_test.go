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
		instance ec2.Instance
		client   *fakes.InstancesClient
		id       *string
		keyName  *string
	)

	BeforeEach(func() {
		client = &fakes.InstancesClient{}
		id = aws.String("the-id")
		keyName = aws.String("the-key-name")
		tags := []*awsec2.Tag{}

		instance = ec2.NewInstance(client, id, keyName, tags)
	})

	Describe("Delete", func() {
		It("terminates the instance", func() {
			err := instance.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.TerminateInstancesCall.CallCount).To(Equal(1))
			Expect(client.TerminateInstancesCall.Receives.Input.InstanceIds).To(HaveLen(1))
			Expect(client.TerminateInstancesCall.Receives.Input.InstanceIds[0]).To(Equal(id))
		})

		Context("the client fails", func() {
			BeforeEach(func() {
				client.TerminateInstancesCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := instance.Delete()
				Expect(err).To(MatchError("FAILED deleting EC2 Instance the-id (KeyPairName:the-key-name): banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(instance.Name()).To(Equal("the-id (KeyPairName:the-key-name)"))
		})
	})

	Describe("Type", func() {
		It("returns \"instance\"", func() {
			Expect(instance.Type()).To(Equal("EC2 Instance"))
		})
	})
})
