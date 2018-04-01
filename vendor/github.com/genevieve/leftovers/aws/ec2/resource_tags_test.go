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

var _ = Describe("ResourceTags", func() {
	var (
		client       *fakes.TagsClient
		resourceTags ec2.ResourceTags
	)

	BeforeEach(func() {
		client = &fakes.TagsClient{}
		resourceTags = ec2.NewResourceTags(client)
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			client.DescribeTagsCall.Returns.Output = &awsec2.DescribeTagsOutput{
				Tags: []*awsec2.TagDescription{{
					ResourceId: aws.String("the-resource-id"),
					Key:        aws.String("the-key"),
					Value:      aws.String("the-value"),
				}},
			}
		})

		It("deletes the resource tags", func() {
			err := resourceTags.Delete("vpc", "vpc-id")
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeTagsCall.CallCount).To(Equal(1))
			Expect(client.DescribeTagsCall.Receives.Input.Filters[0].Name).To(Equal(aws.String("resource-type")))
			Expect(client.DescribeTagsCall.Receives.Input.Filters[0].Values[0]).To(Equal(aws.String("vpc")))
			Expect(client.DescribeTagsCall.Receives.Input.Filters[1].Name).To(Equal(aws.String("resource-id")))
			Expect(client.DescribeTagsCall.Receives.Input.Filters[1].Values[0]).To(Equal(aws.String("vpc-id")))

			Expect(client.DeleteTagsCall.CallCount).To(Equal(1))
			Expect(client.DeleteTagsCall.Receives.Input.Tags[0].Key).To(Equal(aws.String("the-key")))
			Expect(client.DeleteTagsCall.Receives.Input.Tags[0].Value).To(Equal(aws.String("the-value")))
			Expect(client.DeleteTagsCall.Receives.Input.Resources[0]).To(Equal(aws.String("the-resource-id")))
		})

		Context("when the client fails to describe tags", func() {
			BeforeEach(func() {
				client.DescribeTagsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error and does not try deleting them", func() {
				err := resourceTags.Delete("vpc", "vpc-id")
				Expect(err).To(MatchError("Describe tags: some error"))
			})
		})

		Context("when the client fails to delete the tag", func() {
			BeforeEach(func() {
				client.DeleteTagsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				err := resourceTags.Delete("vpc", "vpc-id")
				Expect(err).To(MatchError("Delete the-key:the-value: some error"))
			})
		})
	})
})
