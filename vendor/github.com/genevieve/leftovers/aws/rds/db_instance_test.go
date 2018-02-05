package rds_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevieve/leftovers/aws/rds"
	"github.com/genevieve/leftovers/aws/rds/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DBInstance", func() {
	var (
		dbInstance rds.DBInstance
		client     *fakes.DBInstancesClient
		name       *string
	)

	BeforeEach(func() {
		client = &fakes.DBInstancesClient{}
		name = aws.String("the-name")

		dbInstance = rds.NewDBInstance(client, name)
	})

	Describe("Delete", func() {
		It("deletes the db instance", func() {
			err := dbInstance.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteDBInstanceCall.CallCount).To(Equal(1))
			Expect(client.DeleteDBInstanceCall.Receives.Input.DBInstanceIdentifier).To(Equal(name))
		})

		Context("when the client fails", func() {
			BeforeEach(func() {
				client.DeleteDBInstanceCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := dbInstance.Delete()
				Expect(err).To(MatchError("FAILED deleting db instance the-name: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(dbInstance.Name()).To(Equal("the-name"))
		})
	})
})
