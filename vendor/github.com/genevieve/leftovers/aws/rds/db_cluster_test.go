package rds_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevieve/leftovers/aws/rds"
	"github.com/genevieve/leftovers/aws/rds/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DBCluster", func() {
	var (
		dbCluster    rds.DBCluster
		client       *fakes.DBClustersClient
		name         *string
		skipSnapshot *bool
	)

	BeforeEach(func() {
		client = &fakes.DBClustersClient{}
		name = aws.String("the-name")
		skipSnapshot = aws.Bool(true)

		dbCluster = rds.NewDBCluster(client, name)
	})

	Describe("Delete", func() {
		It("deletes the db cluster", func() {
			err := dbCluster.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteDBClusterCall.CallCount).To(Equal(1))
			Expect(client.DeleteDBClusterCall.Receives.Input.DBClusterIdentifier).To(Equal(name))
			Expect(client.DeleteDBClusterCall.Receives.Input.SkipFinalSnapshot).To(Equal(skipSnapshot))
		})

		Context("when the client fails", func() {
			BeforeEach(func() {
				client.DeleteDBClusterCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := dbCluster.Delete()
				Expect(err).To(MatchError("Delete: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(dbCluster.Name()).To(Equal("the-name"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(dbCluster.Type()).To(Equal("RDS DB Cluster"))
		})
	})
})
