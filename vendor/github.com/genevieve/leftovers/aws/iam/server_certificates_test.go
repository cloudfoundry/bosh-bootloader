package iam_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/genevieve/leftovers/aws/iam"
	"github.com/genevieve/leftovers/aws/iam/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServerCertificates", func() {
	var (
		client *fakes.ServerCertificatesClient
		logger *fakes.Logger

		serverCertificates iam.ServerCertificates
	)

	BeforeEach(func() {
		client = &fakes.ServerCertificatesClient{}
		logger = &fakes.Logger{}

		serverCertificates = iam.NewServerCertificates(client, logger)
	})

	Describe("List", func() {
		var filter string

		BeforeEach(func() {
			logger.PromptCall.Returns.Proceed = true
			client.ListServerCertificatesCall.Returns.Output = &awsiam.ListServerCertificatesOutput{
				ServerCertificateMetadataList: []*awsiam.ServerCertificateMetadata{{
					ServerCertificateName: aws.String("banana-cert"),
				}},
			}
		})

		It("deletes iam server certificates", func() {
			items, err := serverCertificates.List(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListServerCertificatesCall.CallCount).To(Equal(1))

			Expect(logger.PromptCall.CallCount).To(Equal(1))

			Expect(items).To(HaveLen(1))
			// Expect(items).To(HaveKeyWithValue("banana-cert", ""))
		})

		Context("when the client fails to list server certificates", func() {
			BeforeEach(func() {
				client.ListServerCertificatesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := serverCertificates.List(filter)
				Expect(err).To(MatchError("Listing server certificates: some error"))
			})
		})

		Context("when the certificate name does not contain the filter", func() {
			It("does not return it in the list", func() {
				items, err := serverCertificates.List("kiwi")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.CallCount).To(Equal(0))
				Expect(items).To(HaveLen(0))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptCall.Returns.Proceed = false
			})

			It("does not return it in the list", func() {
				items, err := serverCertificates.List(filter)
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptCall.Receives.Message).To(Equal("Are you sure you want to delete server certificate banana-cert?"))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
