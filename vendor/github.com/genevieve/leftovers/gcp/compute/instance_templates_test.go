package compute_test

import (
	"errors"

	"github.com/genevieve/leftovers/gcp/compute"
	"github.com/genevieve/leftovers/gcp/compute/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gcpcompute "google.golang.org/api/compute/v1"
)

var _ = Describe("InstanceTemplates", func() {
	var (
		client *fakes.InstanceTemplatesClient
		logger *fakes.Logger

		instanceTemplates compute.InstanceTemplates
	)

	BeforeEach(func() {
		client = &fakes.InstanceTemplatesClient{}
		logger = &fakes.Logger{}

		instanceTemplates = compute.NewInstanceTemplates(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptWithDetailsCall.Returns.Proceed = true
			client.ListInstanceTemplatesCall.Returns.Output = []*gcpcompute.InstanceTemplate{{
				Name: "banana-template",
			}}
			filter = "banana"
		})

		It("lists, filters, and prompts for instance templates to delete", func() {
			list, err := instanceTemplates.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListInstanceTemplatesCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("Instance Template"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("banana-template"))

			Expect(list).To(HaveLen(1))
		})

		Context("when the client fails to list instance templates", func() {
			BeforeEach(func() {
				client.ListInstanceTemplatesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := instanceTemplates.List(filter)
				Expect(err).To(MatchError("List Instance Templates: some error"))
			})
		})

		Context("when the instance template name does not contain the filter", func() {
			It("does not add it to the list", func() {
				list, err := instanceTemplates.List("grape")
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
				list, err := instanceTemplates.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(list).To(HaveLen(0))
			})
		})
	})
})
