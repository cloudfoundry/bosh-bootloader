package iam_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/multierror"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CertificateValidator", func() {
	Describe("Validate", func() {
		var (
			certificateValidator iam.CertificateValidator
			certFilePath         string
			keyFilePath          string
			chainFilePath        string
			certNonPEMFilePath   string
			keyNonPEMFilePath    string
			chainNonPEMFilePath  string
			otherKeyFilePath     string
			otherCertFilePath    string
			otherChainFilePath   string
		)

		BeforeEach(func() {
			certificateValidator = iam.NewCertificateValidator()
			chainFilePath = "fixtures/bbl-chain.crt"
			certFilePath = "fixtures/bbl.crt"
			keyFilePath = "fixtures/bbl.key"

			otherChainFilePath = "fixtures/other-bbl-chain.crt"
			otherCertFilePath = "fixtures/other-bbl.crt"
			otherKeyFilePath = "fixtures/other-bbl.key"

			createTempFile := func() string {
				file, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())
				defer file.Close()

				return file.Name()
			}

			certNonPEMFilePath = createTempFile()
			keyNonPEMFilePath = createTempFile()
			chainNonPEMFilePath = createTempFile()

			iam.ResetStat()
			iam.ResetReadAll()
		})

		It("does not return an error when cert and key are valid", func() {
			err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, "")

			Expect(err).NotTo(HaveOccurred())
		})

		It("does not return an error when cert, key, and chain are valid", func() {
			err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, chainFilePath)

			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error if cert and key are not provided", func() {
			err := certificateValidator.Validate("some-command-name", "", "", "")
			expectedErr := multierror.NewMultiError("some-command-name")
			expectedErr.Add(errors.New("--cert is required"))
			expectedErr.Add(errors.New("--key is required"))

			Expect(err).To(Equal(expectedErr))
		})

		It("returns an error if the cert key file does not exist", func() {
			err := certificateValidator.Validate("some-command-name", "/some/fake/cert/path", "/some/fake/key/path", "")
			expectedErr := multierror.NewMultiError("some-command-name")
			expectedErr.Add(errors.New(`certificate file not found: "/some/fake/cert/path"`))
			expectedErr.Add(errors.New(`key file not found: "/some/fake/key/path"`))

			Expect(err).To(Equal(expectedErr))
		})

		It("returns an error if the cert and key are not regular files", func() {
			err := certificateValidator.Validate("some-command-name", "/dev/null", "/dev/null", "")
			expectedErr := multierror.NewMultiError("some-command-name")
			expectedErr.Add(errors.New(`certificate is not a regular file: "/dev/null"`))
			expectedErr.Add(errors.New(`key is not a regular file: "/dev/null"`))

			Expect(err).To(Equal(expectedErr))
		})

		It("returns an error if the cert and key are not PEM encoded", func() {
			err := certificateValidator.Validate("some-command-name", certNonPEMFilePath, keyNonPEMFilePath, "")

			expectedErr := multierror.NewMultiError("some-command-name")
			expectedErr.Add(fmt.Errorf(`certificate is not PEM encoded: %q`, certNonPEMFilePath))
			expectedErr.Add(fmt.Errorf(`key is not PEM encoded: %q`, keyNonPEMFilePath))

			Expect(err).To(Equal(expectedErr))
		})

		It("returns an error if the key and cert are not compatible", func() {
			err := certificateValidator.Validate("some-command-name", certFilePath, otherKeyFilePath, "")

			expectedErr := multierror.NewMultiError("some-command-name")
			expectedErr.Add(errors.New("certificate and key mismatch"))
			Expect(err).To(Equal(expectedErr))
		})

		Context("chain is provided", func() {
			It("returns an error when chain file does not exist", func() {
				err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, "/some/fake/chain/path")
				expectedErr := multierror.NewMultiError("some-command-name")
				expectedErr.Add(errors.New(`chain file not found: "/some/fake/chain/path"`))

				Expect(err).To(Equal(expectedErr))
			})

			It("returns an error when chain file is not a regular file", func() {
				err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, "/dev/null")
				expectedErr := multierror.NewMultiError("some-command-name")
				expectedErr.Add(errors.New(`chain is not a regular file: "/dev/null"`))

				Expect(err).To(Equal(expectedErr))
			})

			It("returns an error if the chain is not PEM encoded", func() {
				err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, chainNonPEMFilePath)

				expectedErr := multierror.NewMultiError("some-command-name")
				expectedErr.Add(fmt.Errorf(`chain is not PEM encoded: %q`, chainNonPEMFilePath))

				Expect(err).To(Equal(expectedErr))
			})

			It("returns an error if the chain and cert are not compatible", func() {
				err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, otherChainFilePath)

				expectedErr := multierror.NewMultiError("some-command-name")
				expectedErr.Add(errors.New("certificate and chain mismatch: x509: certificate signed by unknown authority"))
				Expect(err).To(Equal(expectedErr))
			})

			It("returns multiple errors if the cert, key and chain are incompatiable", func() {
				err := certificateValidator.Validate("some-command-name", certFilePath, otherKeyFilePath, otherChainFilePath)
				expectedErr := multierror.NewMultiError("some-command-name")
				expectedErr.Add(errors.New("certificate and key mismatch"))
				expectedErr.Add(errors.New("certificate and chain mismatch: x509: certificate signed by unknown authority"))

				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the certificate, key, and chain cannot be read", func() {
				createTempFile := func() string {
					file, err := ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())
					defer file.Close()

					err = os.Chmod(file.Name(), 0100)
					Expect(err).NotTo(HaveOccurred())

					return file.Name()
				}

				certFile := createTempFile()
				keyFile := createTempFile()
				chainFile := createTempFile()

				err := certificateValidator.Validate("some-command-name", certFile, keyFile, chainFile)
				expectedErr := multierror.NewMultiError("some-command-name")
				expectedErr.Add(fmt.Errorf("open %s: permission denied", certFile))
				expectedErr.Add(fmt.Errorf("open %s: permission denied", keyFile))
				expectedErr.Add(fmt.Errorf("open %s: permission denied", chainFile))

				Expect(err).To(Equal(expectedErr))
			})

			It("returns an error when file info cannot be retrieved", func() {
				iam.SetStat(func(string) (os.FileInfo, error) {
					return nil, errors.New("failed to retrieve file info")
				})

				err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, chainFilePath)

				expectedErr := multierror.NewMultiError("some-command-name")
				expectedErr.Add(fmt.Errorf("failed to retrieve file info: %s", certFilePath))
				expectedErr.Add(fmt.Errorf("failed to retrieve file info: %s", keyFilePath))
				expectedErr.Add(fmt.Errorf("failed to retrieve file info: %s", chainFilePath))

				Expect(err).To(Equal(expectedErr))
			})

			It("returns an error when the file cannot be read", func() {
				iam.SetReadAll(func(io.Reader) ([]byte, error) {
					return []byte{}, errors.New("bad read")
				})

				err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, chainFilePath)

				expectedErr := multierror.NewMultiError("some-command-name")
				expectedErr.Add(fmt.Errorf("bad read: %s", certFilePath))
				expectedErr.Add(fmt.Errorf("bad read: %s", keyFilePath))
				expectedErr.Add(fmt.Errorf("bad read: %s", chainFilePath))

				Expect(err).To(Equal(expectedErr))
			})

			It("returns an error when the private key is not valid rsa", func() {
				file, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())
				defer file.Close()

				err = ioutil.WriteFile(file.Name(), []byte(`
-----BEGIN RSA PRIVATE KEY-----
-----END RSA PRIVATE KEY-----
				`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = certificateValidator.Validate("some-command-name", certFilePath, file.Name(), chainFilePath)
				expectedErr := multierror.NewMultiError("some-command-name")
				expectedErr.Add(errors.New("failed to parse private key: asn1: syntax error: sequence truncated"))

				Expect(err).To(Equal(expectedErr))
			})

			It("returns an error when the certificate is not valid", func() {
				file, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())
				defer file.Close()

				err = ioutil.WriteFile(file.Name(), []byte(`
-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----
				`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = certificateValidator.Validate("some-command-name", file.Name(), keyFilePath, chainFilePath)
				expectedErr := multierror.NewMultiError("some-command-name")
				expectedErr.Add(errors.New("failed to parse certificate: asn1: syntax error: sequence truncated"))

				Expect(err).To(Equal(expectedErr))
			})

			It("returns an error when the chain is not valid", func() {
				file, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())
				defer file.Close()

				err = ioutil.WriteFile(file.Name(), []byte(`
-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----
				`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, file.Name())
				expectedErr := multierror.NewMultiError("some-command-name")
				expectedErr.Add(errors.New("failed to parse chain"))

				Expect(err).To(Equal(expectedErr))
			})
		})
	})
})
