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

var _ = Describe("DBInstances", func() {
	var (
		client *fakes.DBInstancesClient
		logger *fakes.Logger

		dbInstances rds.DBInstances
	)

	BeforeEach(func() {
		client = &fakes.DBInstancesClient{}
		logger = &fakes.Logger{}

		dbInstances = rds.NewDBInstances(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptCall.Returns.Proceed = true
			client.DescribeDBInstancesCall.Returns.Output = &awsrds.DescribeDBInstancesOutput{
				DBInstances: []*awsrds.DBInstance{{
					DBInstanceIdentifier: aws.String("banana"),
				}},
			}
			filter = "ban"
		})

		It("deletes db instances", func() {
			items, err := dbInstances.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeDBInstancesCall.CallCount).To(Equal(1))

			Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete db instance banana?"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the client fails to list db instances", func() {
			BeforeEach(func() {
				client.DescribeDBInstancesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := dbInstances.List(filter)
				Expect(err).To(MatchError("Describing db instances: some error"))
			})
		})

		Context("when the db instance name does not contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := dbInstances.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DescribeDBInstancesCall.CallCount).To(Equal(1))
				Expect(logger.PromptCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptCall.Returns.Proceed = false
			})

			It("does not return it in the list", func() {
				items, err := dbInstances.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete db instance banana?"))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
