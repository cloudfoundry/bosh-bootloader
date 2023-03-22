package certs_test

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/certs"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"
	"github.com/cloudfoundry/multierror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CertificateValidator", func() {
	var (
		certificateValidator certs.Validator
		certFilePath         string
		keyFilePath          string
		certNonPEMFilePath   string
		keyNonPEMFilePath    string
		pkcs12CertFilePath   string
		passwordFilePath     string
	)

	BeforeEach(func() {
		certificateValidator = certs.NewValidator()

		var err error
		certFilePath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
		Expect(err).NotTo(HaveOccurred())

		keyFilePath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
		Expect(err).NotTo(HaveOccurred())

		certNonPEMFilePath, err = testhelpers.WriteContentsToTempFile("not a cert")
		Expect(err).NotTo(HaveOccurred())

		keyNonPEMFilePath, err = testhelpers.WriteContentsToTempFile("not a key")
		Expect(err).NotTo(HaveOccurred())

		pkcs12CertFile, err := base64.StdEncoding.DecodeString(testhelpers.PFX_BASE64)
		Expect(err).NotTo(HaveOccurred())
		pkcs12CertFilePath, err = testhelpers.WriteByteContentsToTempFile(pkcs12CertFile)
		Expect(err).NotTo(HaveOccurred())

		passwordFilePath, err = testhelpers.WriteContentsToTempFile(testhelpers.PFX_PASSWORD)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("ReadPKCS12", func() {
		Context("when cert and password files exist and can be read", func() {
			It("returns cert and password data", func() {
				certData, err := certificateValidator.ReadPKCS12(certNonPEMFilePath, passwordFilePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(certData.Cert)).To(Equal("not a cert"))
				Expect(string(certData.Key)).To(Equal("SuperSecretPassword"))
			})
		})

		Context("if cert and password are not provided", func() {
			It("returns an error", func() {
				_, err := certificateValidator.ReadPKCS12("", "")
				expectedErr := multierror.NewMultiError("")
				expectedErr.Add(errors.New("--lb-cert is required"))
				expectedErr.Add(errors.New("--lb-key is required"))

				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("if the cert password file does not exist", func() {
			It("returns an error", func() {
				_, err := certificateValidator.ReadPKCS12("/some/fake/cert/path", "/some/fake/key/path")
				expectedErr := multierror.NewMultiError("")
				expectedErr.Add(errors.New(`certificate file not found: "/some/fake/cert/path"`))
				expectedErr.Add(errors.New(`key file not found: "/some/fake/key/path"`))

				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("if the cert and password are not regular files", func() {
			It("returns an error", func() {
				_, err := certificateValidator.ReadPKCS12("/dev/null", "/dev/null")
				expectedErr := multierror.NewMultiError("")
				expectedErr.Add(errors.New(`certificate is not a regular file: "/dev/null"`))
				expectedErr.Add(errors.New(`key is not a regular file: "/dev/null"`))

				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("if the password file ends with a newline character", func() {
			It("Strips the newline from the password", func() {
				passwordWithNewlineFilePath, err := testhelpers.WriteContentsToTempFile(fmt.Sprintf("%s\n", testhelpers.PFX_PASSWORD))
				Expect(err).NotTo(HaveOccurred())

				certData, err := certificateValidator.ReadPKCS12(certNonPEMFilePath, passwordWithNewlineFilePath)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(certData.Cert)).To(Equal("not a cert"))
				Expect(string(certData.Key)).To(Equal("SuperSecretPassword"))
			})
		})
	})

	Describe("ReadAndValidatePKCS12", func() {
		Context("when cert and password are valid", func() {
			It("does not return an error", func() {
				_, err := certificateValidator.ReadAndValidatePKCS12(pkcs12CertFilePath, passwordFilePath)

				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("ValidatePKCS12", func() {
		var (
			realCert     []byte
			fakeCert     []byte
			realPassword []byte
			fakePassword []byte
		)

		BeforeEach(func() {
			var err error
			realCert, err = base64.StdEncoding.DecodeString(testhelpers.PFX_BASE64)
			Expect(err).NotTo(HaveOccurred())
			fakeCert = []byte("not a cert")
			realPassword = []byte(testhelpers.PFX_PASSWORD)
			fakePassword = []byte("NotAPassword")
		})

		Context("When the password is correct", func() {
			It("validates successfully", func() {
				err := certificateValidator.ValidatePKCS12(realCert, realPassword)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When the password is incorrect", func() {
			It("returns an error", func() {
				err := certificateValidator.ValidatePKCS12(realCert, fakePassword)
				expectedErr := multierror.NewMultiError("")
				expectedErr.Add(fmt.Errorf("failed to parse certificate: pkcs12: decryption password incorrect"))
				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("When the certificate is invalid", func() {
			It("returns an error", func() {
				err := certificateValidator.ValidatePKCS12(fakeCert, realPassword)
				Expect(err.Error()).To(ContainSubstring("failed to parse certificate: pkcs12:"))
			})
		})
	})

	Describe("ReadAndValidate", func() {
		Context("when using a PKCS#1 key", func() {
			Context("when cert and key are valid", func() {
				It("does not return an error", func() {
					_, err := certificateValidator.ReadAndValidate(certFilePath, keyFilePath)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when cert and key are valid", func() {
				It("does not return an error", func() {
					_, err := certificateValidator.ReadAndValidate(certFilePath, keyFilePath)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("Read", func() {
		Context("when cert and key files exist and can be read", func() {
			It("returns cert and key data", func() {
				certData, err := certificateValidator.Read(certNonPEMFilePath, keyNonPEMFilePath)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(certData.Cert)).To(Equal("not a cert"))
				Expect(string(certData.Key)).To(Equal("not a key"))
			})
		})

		Context("when cert and key files exist and can be read", func() {
			It("returns cert and key data", func() {
				certData, err := certificateValidator.Read(certNonPEMFilePath, keyNonPEMFilePath)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(certData.Cert)).To(Equal("not a cert"))
				Expect(string(certData.Key)).To(Equal("not a key"))
			})
		})

		Context("if cert and key are not provided", func() {
			It("returns an error", func() {
				_, err := certificateValidator.Read("", "")
				expectedErr := multierror.NewMultiError("")
				expectedErr.Add(errors.New("--lb-cert is required"))
				expectedErr.Add(errors.New("--lb-key is required"))

				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("if the cert key file does not exist", func() {
			It("returns an error", func() {
				_, err := certificateValidator.Read("/some/fake/cert/path", "/some/fake/key/path")
				expectedErr := multierror.NewMultiError("")
				expectedErr.Add(errors.New(`certificate file not found: "/some/fake/cert/path"`))
				expectedErr.Add(errors.New(`key file not found: "/some/fake/key/path"`))

				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("if the cert and key are not regular files", func() {
			It("returns an error", func() {
				_, err := certificateValidator.Read("/dev/null", "/dev/null")
				expectedErr := multierror.NewMultiError("")
				expectedErr.Add(errors.New(`certificate is not a regular file: "/dev/null"`))
				expectedErr.Add(errors.New(`key is not a regular file: "/dev/null"`))

				Expect(err).To(Equal(expectedErr))
			})
		})
	})

	Describe("Validate", func() {
		Context("when using a PKCS#1 key", func() {
			var (
				realCert []byte
				realKey  []byte

				otherKey []byte

				fakeCert []byte
				fakeKey  []byte

				invalidKey  []byte
				invalidCert []byte
			)

			BeforeEach(func() {
				realCert = []byte(testhelpers.BBL_CERT)
				realKey = []byte(testhelpers.BBL_KEY)

				otherKey = []byte(testhelpers.OTHER_BBL_KEY)

				fakeCert = []byte("not a cert")
				fakeKey = []byte("not a key")

				invalidKey = []byte("-----BEGIN RSA PRIVATE KEY-----\n-----END RSA PRIVATE KEY-----")
				invalidCert = []byte("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----")
			})

			Context("when cert and key are valid", func() {
				It("does not return an error", func() {
					err := certificateValidator.Validate(realCert, realKey)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("if the cert and key are not PEM encoded", func() {
				It("returns an error", func() {
					err := certificateValidator.Validate(fakeCert, fakeKey)

					expectedErr := multierror.NewMultiError("")
					expectedErr.Add(fmt.Errorf(`certificate is not PEM encoded: %q`, "not a cert"))
					expectedErr.Add(fmt.Errorf(`key is not PEM encoded: %q`, "not a key"))

					Expect(err).To(Equal(expectedErr))
				})
			})

			Context("if the key and cert are not compatible", func() {
				It("returns an error", func() {
					err := certificateValidator.Validate(realCert, otherKey)

					expectedErr := multierror.NewMultiError("")
					expectedErr.Add(errors.New("tls: private key does not match public key"))
					Expect(err).To(Equal(expectedErr))
				})
			})

			Context("when the key is not valid", func() {
				It("returns an error", func() {
					err := certificateValidator.Validate(realCert, invalidKey)

					expectedErr := multierror.NewMultiError("")
					expectedErr.Add(errors.New("tls: failed to parse private key"))
					Expect(err).To(Equal(expectedErr))
				})
			})

			Context("when the cert is not valid", func() {
				It("returns an error", func() {
					err := certificateValidator.Validate(invalidCert, realKey)

					expectedErr := multierror.NewMultiError("")
					expectedErr.Add(errors.New("asn1: syntax error: sequence truncated"))
					Expect(err).To(Equal(expectedErr))
				})
			})
		})

		Context("when using a PKCS#8 key", func() {
			var (
				cert []byte
				key  []byte
			)

			BeforeEach(func() {
				var err error
				cert, err = os.ReadFile("fixtures/pkcs8.crt")
				Expect(err).NotTo(HaveOccurred())

				key, err = os.ReadFile("fixtures/pkcs8.key")
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when cert and key are valid", func() {
				It("does not return an error", func() {
					err := certificateValidator.Validate(cert, key)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
