package iam_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam/fakes"

	awsiam "github.com/aws/aws-sdk-go/service/iam"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CertificateDescriber", func() {
	var (
		iamClient *fakes.Client
		describer iam.CertificateDescriber
	)

	BeforeEach(func() {
		iamClient = &fakes.Client{}
		describer = iam.NewCertificateDescriber(iamClient)
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

			Expect(iamClient.GetServerCertificateArgsForCall(0).ServerCertificateName).To(Equal(aws.String("some-certificate")))

			Expect(certificate.Name).To(Equal("some-certificate"))
			Expect(certificate.Body).To(Equal("some-certificate-body"))
			Expect(certificate.Chain).To(Equal("some-chain-body"))
			Expect(certificate.ARN).To(Equal("some-certificate-arn"))
		})

		Context("failure cases", func() {
			Context("when the ServerCertificate is nil", func() {
				BeforeEach(func() {
					iamClient.GetServerCertificateReturns(&awsiam.GetServerCertificateOutput{
						ServerCertificate: nil,
					}, nil)
				})

				It("returns an error", func() {
					_, err := describer.Describe("some-certificate")
					Expect(err).To(MatchError(iam.CertificateDescriptionFailure))
				})
			})

			Context("when the ServerCertificateMetadata is nil", func() {
				BeforeEach(func() {
					iamClient.GetServerCertificateReturns(&awsiam.GetServerCertificateOutput{
						ServerCertificate: &awsiam.ServerCertificate{
							ServerCertificateMetadata: nil,
						},
					}, nil)
				})

				It("returns an error", func() {
					_, err := describer.Describe("some-certificate")
					Expect(err).To(MatchError(iam.CertificateDescriptionFailure))
				})
			})

			Context("when the certificate cannot be described", func() {
				BeforeEach(func() {
					iamClient.GetServerCertificateReturns(nil, awserr.NewRequestFailure(
						awserr.New("boom",
							"something bad happened",
							errors.New(""),
						), 404, "0",
					))
				})

				It("returns an error", func() {
					_, err := describer.Describe("some-certificate")
					Expect(err).To(MatchError(ContainSubstring("something bad happened")))
				})
			})

			Context("when the certificate does not exist", func() {
				BeforeEach(func() {
					iamClient.GetServerCertificateReturns(nil, awserr.NewRequestFailure(
						awserr.New("NoSuchEntity",
							"The Server Certificate with name some-certificate cannot be found.",
							errors.New(""),
						), 404, "0",
					))
				})

				It("returns a CertificateNotFound error ", func() {
					_, err := describer.Describe("some-certificate")
					Expect(err).To(MatchError(iam.CertificateNotFound))
				})
			})
		})
	})
})
