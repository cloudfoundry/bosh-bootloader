package sql_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/sql"
	"github.com/genevieve/leftovers/gcp/sql/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Instance", func() {
	var (
		client *fakes.InstancesClient
		name   string

		instance sql.Instance
	)

	BeforeEach(func() {
		client = &fakes.InstancesClient{}
		name = "banana"

		instance = sql.NewInstance(client, name)
	})

	Describe("Delete", func() {
		It("deletes the instance", func() {
			err := instance.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteInstanceCall.CallCount).To(Equal(1))
			Expect(client.DeleteInstanceCall.Receives.Instance).To(Equal(name))
		})

		Context("when the client fails to delete the instance", func() {
			BeforeEach(func() {
				client.DeleteInstanceCall.Returns.Error = errors.New("the-error")
			})

			It("returns the error", func() {
				err := instance.Delete()
				Expect(err).To(MatchError("Delete: the-error"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the name", func() {
			Expect(instance.Name()).To(Equal(name))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(instance.Type()).To(Equal("SQL Instance"))
		})
	})
})
