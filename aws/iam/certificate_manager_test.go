package iam_test

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CertificateManager", func() {
	var (
		iamClient            *fakes.IAMClient
		certificateUploader  *fakes.CertificateUploader
		certificateDescriber *fakes.CertificateDescriber
		certificateDeleter   *fakes.CertificateDeleter
		manager              iam.CertificateManager
		certificateFile      *os.File
		privateKeyFile       *os.File
	)

	BeforeEach(func() {
		var err error
		iamClient = &fakes.IAMClient{}
		certificateUploader = &fakes.CertificateUploader{}
		certificateDescriber = &fakes.CertificateDescriber{}
		certificateDeleter = &fakes.CertificateDeleter{}
		manager = iam.NewCertificateManager(certificateUploader, certificateDescriber, certificateDeleter)

		certificateFile, err = ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())

		privateKeyFile, err = ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())

	})

	Describe("CreateOrUpdate", func() {
		It("creates the certificate if it doesn't exist", func() {
			certificateDescriber.DescribeCall.Returns.Certificate = iam.Certificate{}
			certificateDescriber.DescribeCall.Returns.Error = iam.CertificateNotFound
			certificateUploader.UploadCall.Returns.CertificateName = "some-new-certificate"

			certificateName, err := manager.CreateOrUpdate("some-non-existant-certificate", certificateFile.Name(), privateKeyFile.Name(), iamClient)
			Expect(err).NotTo(HaveOccurred())

			Expect(certificateName).To(Equal("some-new-certificate"))

			Expect(certificateDescriber.DescribeCall.CallCount).To(Equal(1))
			Expect(certificateDescriber.DescribeCall.Receives.CertificateName).To(Equal("some-non-existant-certificate"))
			Expect(certificateDescriber.DescribeCall.Receives.IAMClient).To(Equal(iamClient))

			Expect(certificateUploader.UploadCall.CallCount).To(Equal(1))
			Expect(certificateUploader.UploadCall.Receives.CertificatePath).To(Equal(certificateFile.Name()))
			Expect(certificateUploader.UploadCall.Receives.PrivateKeyPath).To(Equal(privateKeyFile.Name()))
			Expect(certificateUploader.UploadCall.Receives.IAMClient).To(Equal(iamClient))

			Expect(certificateDeleter.DeleteCall.CallCount).To(Equal(0))
		})

		It("does not upload a new certificate if the certificate is the same ignoring leading and trailing whitespace", func() {
			err := ioutil.WriteFile(certificateFile.Name(), []byte("   some-certificate-body\n   "), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(privateKeyFile.Name(), []byte("some-private-key-body"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			certificateDescriber.DescribeCall.Returns.Certificate = iam.Certificate{
				Name: "some-certificate",
				Body: "some-certificate-body",
			}

			certificateName, err := manager.CreateOrUpdate("some-certificate", certificateFile.Name(), privateKeyFile.Name(), iamClient)
			Expect(err).NotTo(HaveOccurred())

			Expect(certificateName).To(Equal("some-certificate"))

			Expect(certificateDescriber.DescribeCall.CallCount).To(Equal(1))
			Expect(certificateDescriber.DescribeCall.Receives.CertificateName).To(Equal("some-certificate"))
			Expect(certificateDescriber.DescribeCall.Receives.IAMClient).To(Equal(iamClient))

			Expect(certificateDeleter.DeleteCall.CallCount).To(Equal(0))

			Expect(certificateUploader.UploadCall.CallCount).To(Equal(0))
		})

		It("uploads new certificate if the certificate is different", func() {
			err := ioutil.WriteFile(certificateFile.Name(), []byte("some-other-certificate-body"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(privateKeyFile.Name(), []byte("some-other-private-key-body"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			certificateDescriber.DescribeCall.Returns.Certificate = iam.Certificate{
				Name: "some-certificate",
				Body: "some-certificate-body",
			}

			certificateUploader.UploadCall.Returns.CertificateName = "some-new-certificate"

			certificateName, err := manager.CreateOrUpdate("some-certificate", certificateFile.Name(), privateKeyFile.Name(), iamClient)
			Expect(err).NotTo(HaveOccurred())

			Expect(certificateName).To(Equal("some-new-certificate"))

			Expect(certificateDescriber.DescribeCall.CallCount).To(Equal(1))
			Expect(certificateDescriber.DescribeCall.Receives.CertificateName).To(Equal("some-certificate"))
			Expect(certificateDescriber.DescribeCall.Receives.IAMClient).To(Equal(iamClient))

			Expect(certificateDeleter.DeleteCall.CallCount).To(Equal(1))
			Expect(certificateDeleter.DeleteCall.Receives.CertificateName).To(Equal("some-certificate"))
			Expect(certificateDeleter.DeleteCall.Receives.IAMClient).To(Equal(iamClient))

			Expect(certificateUploader.UploadCall.CallCount).To(Equal(1))
			Expect(certificateUploader.UploadCall.Receives.CertificatePath).To(Equal(certificateFile.Name()))
			Expect(certificateUploader.UploadCall.Receives.PrivateKeyPath).To(Equal(privateKeyFile.Name()))
			Expect(certificateUploader.UploadCall.Receives.IAMClient).To(Equal(iamClient))
		})

		It("doesn't call describe when no existing certificate name is passed in", func() {
			_, err := manager.CreateOrUpdate("", certificateFile.Name(), privateKeyFile.Name(), iamClient)
			Expect(err).NotTo(HaveOccurred())
			Expect(certificateDescriber.DescribeCall.CallCount).To(Equal(0))
			Expect(certificateUploader.UploadCall.CallCount).To(Equal(1))
		})

		Context("failure cases", func() {
			It("returns an error if describe returns an error that is not CertificateNotFound", func() {
				certificateDescriber.DescribeCall.Returns.Certificate = iam.Certificate{}
				certificateDescriber.DescribeCall.Returns.Error = errors.New("failed to describe")

				_, err := manager.CreateOrUpdate("some-certificate", certificateFile.Name(), privateKeyFile.Name(), iamClient)
				Expect(err).To(MatchError("failed to describe"))
			})

			It("returns an error if certificate file cannot be read", func() {
				certificateDescriber.DescribeCall.Returns.Certificate = iam.Certificate{}
				certificateDescriber.DescribeCall.Returns.Error = nil

				_, err := manager.CreateOrUpdate("some-certificate", "/path/to/non-existitent/file", privateKeyFile.Name(), iamClient)
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})

			It("returns an error if certificate upload fails", func() {
				certificateDescriber.DescribeCall.Returns.Certificate = iam.Certificate{}
				certificateDescriber.DescribeCall.Returns.Error = iam.CertificateNotFound

				certificateUploader.UploadCall.Returns.Error = errors.New("failed to upload")

				_, err := manager.CreateOrUpdate("some-certificate", certificateFile.Name(), privateKeyFile.Name(), iamClient)
				Expect(err).To(MatchError("failed to upload"))
			})

			It("returns an error if certificate deletion fails", func() {
				err := ioutil.WriteFile(certificateFile.Name(), []byte("some-other-certificate-body"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				certificateDescriber.DescribeCall.Returns.Certificate = iam.Certificate{
					Name: "some-certificate",
					Body: "some-certificate-body",
				}

				certificateDeleter.DeleteCall.Returns.Error = errors.New("deletion failed")

				_, err = manager.CreateOrUpdate("some-certificate", certificateFile.Name(), privateKeyFile.Name(), iamClient)
				Expect(err).To(MatchError("deletion failed"))
			})

			It("returns an error if certificate upload fails", func() {
				err := ioutil.WriteFile(certificateFile.Name(), []byte("some-other-certificate-body"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				certificateDescriber.DescribeCall.Returns.Certificate = iam.Certificate{
					Name: "some-certificate",
					Body: "some-certificate-body",
				}

				certificateUploader.UploadCall.Returns.Error = errors.New("upload failed")

				_, err = manager.CreateOrUpdate("some-certificate", certificateFile.Name(), privateKeyFile.Name(), iamClient)
				Expect(err).To(MatchError("upload failed"))
			})
		})
	})

	Describe("Create", func() {
		It("creates the given certificate", func() {
			certificateUploader.UploadCall.Returns.CertificateName = "some-new-certificate"

			certificateName, err := manager.Create(certificateFile.Name(), privateKeyFile.Name(), iamClient)
			Expect(err).NotTo(HaveOccurred())

			Expect(certificateUploader.UploadCall.CallCount).To(Equal(1))
			Expect(certificateUploader.UploadCall.Receives.CertificatePath).To(Equal(certificateFile.Name()))
			Expect(certificateUploader.UploadCall.Receives.PrivateKeyPath).To(Equal(privateKeyFile.Name()))
			Expect(certificateUploader.UploadCall.Receives.IAMClient).To(Equal(iamClient))

			Expect(certificateName).To(Equal("some-new-certificate"))
		})

		Context("failure cases", func() {
			Context("when certificate uploader fails to upload", func() {
				It("returns an error", func() {
					certificateUploader.UploadCall.Returns.Error = errors.New("upload failed")

					_, err := manager.Create(certificateFile.Name(), privateKeyFile.Name(), iamClient)
					Expect(err).To(MatchError("upload failed"))
				})
			})
		})
	})

	Describe("Delete", func() {
		It("deletes the given certificate", func() {
			err := manager.Delete("some-certificate-name", iamClient)
			Expect(err).NotTo(HaveOccurred())

			Expect(certificateDeleter.DeleteCall.Receives.CertificateName).To(Equal("some-certificate-name"))
			Expect(certificateDeleter.DeleteCall.Receives.IAMClient).To(Equal(iamClient))
		})

		Context("failure cases", func() {
			It("returns an error when certificate fails to delete", func() {
				certificateDeleter.DeleteCall.Returns.Error = errors.New("unknown certificate error")

				err := manager.Delete("some-non-existant-certificate", iamClient)
				Expect(err).To(MatchError("unknown certificate error"))
			})
		})
	})

	Describe("Describe", func() {
		It("returns a certificate", func() {
			certificateDescriber.DescribeCall.Returns.Certificate = iam.Certificate{
				Name: "some-certificate-name",
				ARN:  "some-certificate-arn",
				Body: "some-certificate-body",
			}

			certificate, err := manager.Describe("some-certificate-name", iamClient)
			Expect(err).NotTo(HaveOccurred())

			Expect(certificate).To(Equal(iam.Certificate{
				Name: "some-certificate-name",
				ARN:  "some-certificate-arn",
				Body: "some-certificate-body",
			}))
			Expect(certificateDescriber.DescribeCall.Receives.CertificateName).To(Equal("some-certificate-name"))
			Expect(certificateDescriber.DescribeCall.Receives.IAMClient).To(Equal(iamClient))
		})

		Context("failure cases", func() {
			It("returns an error when the describe fails", func() {
				certificateDescriber.DescribeCall.Returns.Error = errors.New("unknown certificate error")

				_, err := manager.Describe("some-non-existant-certificate", iamClient)
				Expect(err).To(MatchError("unknown certificate error"))
			})
		})
	})
})
