package iam_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CertificateDeleter", func() {
	var (
		iamClient *fakes.Client
		deleter   iam.CertificateDeleter
	)

	BeforeEach(func() {
		iamClient = &fakes.Client{}
		deleter = iam.NewCertificateDeleter(iamClient)
	})

	Describe("Delete", func() {
		It("deletes the certificates with the given name", func() {
			iamClient.DeleteServerCertificateReturns(&awsiam.DeleteServerCertificateOutput{}, nil)

			err := deleter.Delete("some-certificate")
			Expect(err).NotTo(HaveOccurred())

			Expect(iamClient.DeleteServerCertificateArgsForCall(0).ServerCertificateName).To(Equal(aws.String("some-certificate")))
		})

		Context("failure cases", func() {
			It("returns an error when it fails to delete", func() {
				iamClient.DeleteServerCertificateReturns(nil, errors.New("failed to delete certificate"))

				err := deleter.Delete("some-certificate")
				Expect(err).To(MatchError("failed to delete certificate"))
			})
		})
	})
})
