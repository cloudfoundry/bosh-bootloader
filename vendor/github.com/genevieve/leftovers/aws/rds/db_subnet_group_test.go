package rds_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevieve/leftovers/aws/rds"
	"github.com/genevieve/leftovers/aws/rds/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DBSubnetGroup", func() {
	var (
		dbSubnetGroup rds.DBSubnetGroup
		client        *fakes.DBSubnetGroupsClient
		name          *string
	)

	BeforeEach(func() {
		client = &fakes.DBSubnetGroupsClient{}
		name = aws.String("the-name")

		dbSubnetGroup = rds.NewDBSubnetGroup(client, name)
	})

	Describe("Delete", func() {
		It("deletes the db instance", func() {
			err := dbSubnetGroup.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteDBSubnetGroupCall.CallCount).To(Equal(1))
			Expect(client.DeleteDBSubnetGroupCall.Receives.Input.DBSubnetGroupName).To(Equal(name))
		})

		Context("when the client fails", func() {
			BeforeEach(func() {
				client.DeleteDBSubnetGroupCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := dbSubnetGroup.Delete()
				Expect(err).To(MatchError("FAILED deleting db subnet group the-name: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(dbSubnetGroup.Name()).To(Equal("the-name"))
		})
	})
})
