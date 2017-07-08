package iam_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam/fakes"
	awsClientFake "github.com/cloudfoundry/bosh-bootloader/fakes"

	awsiam "github.com/aws/aws-sdk-go/service/iam"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CertificateDescriber", func() {
	var (
		iamClient         *fakes.Client
		awsClientProvider *awsClientFake.AWSClientProvider
		describer         iam.CertificateDescriber
	)

	BeforeEach(func() {
		iamClient = &fakes.Client{}
		awsClientProvider = &awsClientFake.AWSClientProvider{}
		awsClientProvider.GetIAMClientCall.Returns.IAMClient = iamClient

		describer = iam.NewCertificateDescriber(awsClientProvider)
	})

	Describe("Describe", func() {
		It("describes the certificate with the given name", func() {
			iamClient.GetServerCertificateReturns(&awsiam.GetServerCertificateOutput{
				ServerCertificate: &awsiam.ServerCertificate{
					CertificateBody:  aws.String("some-certificate-body"),
					CertificateChain: aws.String("some-chain-body"),
					ServerCertificateMetadata: &awsiam.ServerCertificateMetadata{
						Path:                  aws.String("some-certificate-path"),
						Arn:                   aws.String("some-certificate-arn"),
						ServerCertificateId:   aws.String("some-server-certificate-id"),
						ServerCertificateName: aws.String("some-certificate"),
					},
				},
			}, nil)

			certificate, err := describer.Describe("some-certificate")
			Expect(err).NotTo(HaveOccurred())

			Expect(awsClientProvider.GetIAMClientCall.CallCount).To(Equal(1))

			Expect(iamClient.GetServerCertificateArgsForCall(0).ServerCertificateName).To(Equal(aws.String("some-certificate")))

			Expect(certificate.Name).To(Equal("some-certificate"))
			Expect(certificate.Body).To(Equal("some-certificate-body"))
			Expect(certificate.Chain).To(Equal("some-chain-body"))
			Expect(certificate.ARN).To(Equal("some-certificate-arn"))
		})

		Context("failure cases", func() {
			It("returns an error when the ServerCertificate is nil", func() {
				iamClient.GetServerCertificateReturns(&awsiam.GetServerCertificateOutput{
					ServerCertificate: nil,
				}, nil)

				_, err := describer.Describe("some-certificate")
				Expect(err).To(MatchError(iam.CertificateDescriptionFailure))
			})

			It("returns an error when the ServerCertificateMetadata is nil", func() {
				iamClient.GetServerCertificateReturns(&awsiam.GetServerCertificateOutput{
					ServerCertificate: &awsiam.ServerCertificate{
						ServerCertificateMetadata: nil,
					},
				}, nil)

				_, err := describer.Describe("some-certificate")
				Expect(err).To(MatchError(iam.CertificateDescriptionFailure))
			})

			It("returns an error when the certificate cannot be described", func() {
				iamClient.GetServerCertificateReturns(nil, awserr.NewRequestFailure(
					awserr.New("boom",
						"something bad happened",
						errors.New(""),
					), 404, "0",
				))

				_, err := describer.Describe("some-certificate")
				Expect(err).To(MatchError(ContainSubstring("something bad happened")))
			})

			It("returns a CertificateNotFound error when the certificate does not exist", func() {
				iamClient.GetServerCertificateReturns(nil, awserr.NewRequestFailure(
					awserr.New("NoSuchEntity",
						"The Server Certificate with name some-certificate cannot be found.",
						errors.New(""),
					), 404, "0",
				))

				_, err := describer.Describe("some-certificate")
				Expect(err).To(MatchError(iam.CertificateNotFound))
			})
		})
	})
})
