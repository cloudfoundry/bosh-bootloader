package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/ec2/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Instance", func() {
	var (
		client       *fakes.InstancesClient
		logger       *fakes.Logger
		resourceTags *fakes.ResourceTags
		id           *string
		keyName      *string

		instance ec2.Instance
	)

	BeforeEach(func() {
		client = &fakes.InstancesClient{}
		logger = &fakes.Logger{}
		resourceTags = &fakes.ResourceTags{}
		id = aws.String("the-id")
		keyName = aws.String("the-key-name")
		tags := []*awsec2.Tag{}

		instance = ec2.NewInstance(client, logger, resourceTags, id, keyName, tags)
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			client.DescribeAddressesCall.Returns.Output = &awsec2.DescribeAddressesOutput{
				Addresses: []*awsec2.Address{{
					AllocationId: aws.String("the-allocation-id"),
				}},
			}
			client.DescribeInstancesCall.Returns.Output = &awsec2.DescribeInstancesOutput{
				Reservations: []*awsec2.Reservation{{
					Instances: []*awsec2.Instance{{
						State: &awsec2.InstanceState{Name: aws.String("terminated")},
					}},
				}},
			}
		})

		It("terminates the instance, deletes it's tags, and releases the address", func() {
			err := instance.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeAddressesCall.CallCount).To(Equal(1))
			Expect(client.DescribeAddressesCall.Receives.Input.Filters[0].Name).To(Equal(aws.String("instance-id")))
			Expect(client.DescribeAddressesCall.Receives.Input.Filters[0].Values[0]).To(Equal(id))

			Expect(client.TerminateInstancesCall.CallCount).To(Equal(1))
			Expect(client.TerminateInstancesCall.Receives.Input.InstanceIds).To(HaveLen(1))
			Expect(client.TerminateInstancesCall.Receives.Input.InstanceIds[0]).To(Equal(id))

			Expect(resourceTags.DeleteCall.CallCount).To(Equal(1))
			Expect(resourceTags.DeleteCall.Receives.ResourceType).To(Equal("instance"))
			Expect(resourceTags.DeleteCall.Receives.ResourceId).To(Equal("the-id"))

			Expect(client.ReleaseAddressCall.CallCount).To(Equal(1))
			Expect(client.ReleaseAddressCall.Receives.Input.AllocationId).To(Equal(aws.String("the-allocation-id")))
		})

		Context("when the client fails to describe addresses", func() {
			BeforeEach(func() {
				client.DescribeAddressesCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := instance.Delete()
				Expect(err).To(MatchError("Describe addresses: banana"))
			})
		})

		Context("when the client fails to terminate the instance", func() {
			BeforeEach(func() {
				client.TerminateInstancesCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := instance.Delete()
				Expect(err).To(MatchError("Terminate: banana"))
			})
		})

		Context("when the instance is not found", func() {
			BeforeEach(func() {
				client.TerminateInstancesCall.Returns.Error = awserr.New("InvalidInstanceID.NotFound", "", nil)
			})

			It("returns the error", func() {
				err := instance.Delete()
				Expect(err).NotTo(HaveOccurred())
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

		Context("when the client fails to release the address", func() {
			BeforeEach(func() {
				client.ReleaseAddressCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := instance.Delete()
				Expect(err).To(MatchError("Release address: banana"))
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
