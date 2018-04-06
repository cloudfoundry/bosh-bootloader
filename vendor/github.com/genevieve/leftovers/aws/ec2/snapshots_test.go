package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	awssts "github.com/aws/aws-sdk-go/service/sts"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/ec2/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Snapshots", func() {
	var (
		client    *fakes.SnapshotsClient
		stsClient *fakes.StsClient
		logger    *fakes.Logger

		snapshots ec2.Snapshots
	)

	BeforeEach(func() {
		client = &fakes.SnapshotsClient{}
		stsClient = &fakes.StsClient{}
		logger = &fakes.Logger{}
		logger.PromptWithDetailsCall.Returns.Proceed = true

		snapshots = ec2.NewSnapshots(client, stsClient, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			filter = "snap"
			client.DescribeSnapshotsCall.Returns.Output = &awsec2.DescribeSnapshotsOutput{
				Snapshots: []*awsec2.Snapshot{{
					SnapshotId: aws.String("the-snapshot-id"),
				}},
			}
			stsClient.GetCallerIdentityCall.Returns.Output = &awssts.GetCallerIdentityOutput{
				Account: aws.String("the-account-id"),
			}
		})

		It("returns a list of ec2 snapshots to delete", func() {
			items, err := snapshots.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(stsClient.GetCallerIdentityCall.CallCount).To(Equal(1))

			Expect(client.DescribeSnapshotsCall.CallCount).To(Equal(1))
			Expect(client.DescribeSnapshotsCall.Receives.Input.OwnerIds[0]).To(Equal(aws.String("the-account-id")))
			Expect(client.DescribeSnapshotsCall.Receives.Input.Filters[0].Name).To(Equal(aws.String("status")))
			Expect(client.DescribeSnapshotsCall.Receives.Input.Filters[0].Values[0]).To(Equal(aws.String("completed")))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("EC2 Snapshot"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("the-snapshot-id"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the sts client fails to get the caller identity", func() {
			BeforeEach(func() {
				stsClient.GetCallerIdentityCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := snapshots.List(filter)
				Expect(err).To(MatchError("Get caller identity: some error"))
			})
		})

		Context("when the client fails to describe snapshots", func() {
			BeforeEach(func() {
				client.DescribeSnapshotsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := snapshots.List(filter)
				Expect(err).To(MatchError("Describe EC2 Snapshots: some error"))
			})
		})

		Context("when the snapshot name does not contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := snapshots.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DescribeSnapshotsCall.CallCount).To(Equal(1))
				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))

				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not return it to the list", func() {
				items, err := snapshots.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
