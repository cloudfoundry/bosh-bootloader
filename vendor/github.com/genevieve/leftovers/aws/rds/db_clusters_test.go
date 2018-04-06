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

var _ = Describe("DBClusters", func() {
	var (
		client *fakes.DBClustersClient
		logger *fakes.Logger

		dbClusters rds.DBClusters
	)

	BeforeEach(func() {
		client = &fakes.DBClustersClient{}
		logger = &fakes.Logger{}

		dbClusters = rds.NewDBClusters(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.DescribeDBClustersCall.Returns.Output = &awsrds.DescribeDBClustersOutput{
				DBClusters: []*awsrds.DBCluster{{
					DBClusterIdentifier: aws.String("banana"),
					Status:              aws.String("status"),
				}},
			}
			filter = "ban"
		})

		It("deletes db clusters", func() {
			items, err := dbClusters.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeDBClustersCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("RDS DB Cluster"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the client fails to list db clusters", func() {
			BeforeEach(func() {
				client.DescribeDBClustersCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := dbClusters.List(filter)
				Expect(err).To(MatchError("Describing RDS DB Clusters: some error"))
			})
		})

		Context("when the db cluster name does not contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := dbClusters.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DescribeDBClustersCall.CallCount).To(Equal(1))
				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the db cluster is being deleted", func() {
			BeforeEach(func() {
				client.DescribeDBClustersCall.Returns.Output = &awsrds.DescribeDBClustersOutput{
					DBClusters: []*awsrds.DBCluster{{
						DBClusterIdentifier: aws.String("banana"),
						Status:              aws.String("deleting"),
					}},
				}
			})

			It("does not return it in the list", func() {
				items, err := dbClusters.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not return it in the list", func() {
				items, err := dbClusters.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
