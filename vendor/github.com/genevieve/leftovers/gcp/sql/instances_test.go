package sql_test

import (
	"errors"

	gcpsql "google.golang.org/api/sqladmin/v1beta4"

	"github.com/genevieve/leftovers/gcp/sql"
	"github.com/genevieve/leftovers/gcp/sql/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Instances", func() {
	var (
		client *fakes.InstancesClient
		logger *fakes.Logger

		instances sql.Instances
	)

	BeforeEach(func() {
		client = &fakes.InstancesClient{}
		logger = &fakes.Logger{}

		logger.PromptWithDetailsCall.Returns.Proceed = true

		instances = sql.NewInstances(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			client.ListInstancesCall.Returns.Output = &gcpsql.InstancesListResponse{
				Items: []*gcpsql.DatabaseInstance{{
					Name: "banana-instance",
				}},
			}
			filter = "banana"
		})

		It("lists, filters, and prompts for instances to delete", func() {
			list, err := instances.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListInstancesCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("SQL Instance"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-instance"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list instances", func() {
			BeforeEach(func() {
				client.ListInstancesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := instances.List(filter)
				Expect(err).To(MatchError("List SQL Instances: some error"))
			})
		})

		Context("when the instance name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := instances.List("grape")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(0))
				Expect(list).To(HaveLen(0))
			})
		})

		Context("when the user says no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not add it to the list", func() {
				list, err := instances.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
