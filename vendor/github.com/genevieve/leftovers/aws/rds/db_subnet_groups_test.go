package rds_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsrds "github.com/aws/aws-sdk-go/service/rds"
	"github.com/genevieve/leftovers/aws/rds"
	"github.com/genevieve/leftovers/aws/rds/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DBSubnetGroups", func() {
	var (
		client *fakes.DBSubnetGroupsClient
		logger *fakes.Logger

		dbSubnetGroups rds.DBSubnetGroups
	)

	BeforeEach(func() {
		client = &fakes.DBSubnetGroupsClient{}
		logger = &fakes.Logger{}

		dbSubnetGroups = rds.NewDBSubnetGroups(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptCall.Returns.Proceed = true
			client.DescribeDBSubnetGroupsCall.Returns.Output = &awsrds.DescribeDBSubnetGroupsOutput{
				DBSubnetGroups: []*awsrds.DBSubnetGroup{{
					DBSubnetGroupName: aws.String("banana"),
				}},
			}
			filter = "ban"
		})

		It("deletes db subnet groups", func() {
			items, err := dbSubnetGroups.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeDBSubnetGroupsCall.CallCount).To(Equal(1))

			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete db subnet group banana?"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the client fails to list db subnet groups", func() {
			BeforeEach(func() {
				client.DescribeDBSubnetGroupsCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := dbSubnetGroups.List(filter)
				Expect(err).To(MatchError("Describing db subnet groups: some error"))
			})
		})

		Context("when the db subnet group name does not contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := dbSubnetGroups.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DescribeDBSubnetGroupsCall.CallCount).To(Equal(1))
				Expect(logger.PromptCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptCall.Returns.Proceed = false
			})

			It("does not return it in the list", func() {
				items, err := dbSubnetGroups.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete db subnet group banana?"))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
