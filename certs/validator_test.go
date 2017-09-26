package certs_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/certs"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"
	"github.com/cloudfoundry/multierror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CertificateValidator", func() {
	Describe("Validate", func() {
		var (
			certificateValidator certs.Validator
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
			var err error
			certificateValidator = certs.NewValidator()
			chainFilePath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_CHAIN)
			Expect(err).NotTo(HaveOccurred())

			certFilePath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
			Expect(err).NotTo(HaveOccurred())

			keyFilePath, err = testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
			Expect(err).NotTo(HaveOccurred())

			otherChainFilePath, err = testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_CHAIN)
			Expect(err).NotTo(HaveOccurred())

			otherCertFilePath, err = testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_CERT)
			Expect(err).NotTo(HaveOccurred())

			otherKeyFilePath, err = testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_KEY)
			Expect(err).NotTo(HaveOccurred())

			certNonPEMFilePath, err = testhelpers.WriteContentsToTempFile("")
			Expect(err).NotTo(HaveOccurred())

			keyNonPEMFilePath, err = testhelpers.WriteContentsToTempFile("")
			Expect(err).NotTo(HaveOccurred())

			chainNonPEMFilePath, err = testhelpers.WriteContentsToTempFile("")
			Expect(err).NotTo(HaveOccurred())

			certs.ResetStat()
			certs.ResetReadAll()
		})

		Context("when using a PKCS#1 key", func() {
			Context("when cert and key are valid", func() {
				It("does not return an error", func() {
					err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, "")

					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when cert, key, and chain are valid", func() {
				It("does not return an error", func() {
					err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, chainFilePath)

					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("if cert and key are not provided", func() {
				It("returns an error", func() {
					err := certificateValidator.Validate("some-command-name", "", "", "")
					expectedErr := multierror.NewMultiError("some-command-name")
					expectedErr.Add(errors.New("--cert is required"))
					expectedErr.Add(errors.New("--key is required"))

					Expect(err).To(Equal(expectedErr))
				})
			})

			Context("if the cert key file does not exist", func() {
				It("returns an error", func() {
					err := certificateValidator.Validate("some-command-name", "/some/fake/cert/path", "/some/fake/key/path", "")
					expectedErr := multierror.NewMultiError("some-command-name")
					expectedErr.Add(errors.New(`certificate file not found: "/some/fake/cert/path"`))
					expectedErr.Add(errors.New(`key file not found: "/some/fake/key/path"`))

					Expect(err).To(Equal(expectedErr))
				})
			})

			Context("if the cert and key are not regular files", func() {
				It("returns an error", func() {
					err := certificateValidator.Validate("some-command-name", "/dev/null", "/dev/null", "")
					expectedErr := multierror.NewMultiError("some-command-name")
					expectedErr.Add(errors.New(`certificate is not a regular file: "/dev/null"`))
					expectedErr.Add(errors.New(`key is not a regular file: "/dev/null"`))

					Expect(err).To(Equal(expectedErr))
				})
			})

			Context("if the cert and key are not PEM encoded", func() {
				It("returns an error", func() {
					err := certificateValidator.Validate("some-command-name", certNonPEMFilePath, keyNonPEMFilePath, "")

					expectedErr := multierror.NewMultiError("some-command-name")
					expectedErr.Add(fmt.Errorf(`certificate is not PEM encoded: %q`, certNonPEMFilePath))
					expectedErr.Add(fmt.Errorf(`key is not PEM encoded: %q`, keyNonPEMFilePath))

					Expect(err).To(Equal(expectedErr))
				})
			})

			Context("if the key and cert are not compatible", func() {
				It("returns an error", func() {
					err := certificateValidator.Validate("some-command-name", certFilePath, otherKeyFilePath, "")

					expectedErr := multierror.NewMultiError("some-command-name")
					expectedErr.Add(errors.New("tls: private key does not match public key"))
					Expect(err).To(Equal(expectedErr))
				})
			})

			Context("chain is provided", func() {
				Context("when chain file does not exist", func() {
					It("returns an error", func() {
						err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, "/some/fake/chain/path")
						expectedErr := multierror.NewMultiError("some-command-name")
						expectedErr.Add(errors.New(`chain file not found: "/some/fake/chain/path"`))

						Expect(err).To(Equal(expectedErr))
					})
				})

				Context("when chain file is not a regular file", func() {
					It("returns an error", func() {
						err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, "/dev/null")
						expectedErr := multierror.NewMultiError("some-command-name")
						expectedErr.Add(errors.New(`chain is not a regular file: "/dev/null"`))

						Expect(err).To(Equal(expectedErr))
					})
				})

				Context("if the chain is not PEM encoded", func() {
					It("returns an error", func() {
						err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, chainNonPEMFilePath)

						expectedErr := multierror.NewMultiError("some-command-name")
						expectedErr.Add(fmt.Errorf(`chain is not PEM encoded: %q`, chainNonPEMFilePath))

						Expect(err).To(Equal(expectedErr))
					})
				})

				Context("if the chain and cert are not compatible", func() {
					It("returns an error", func() {
						err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, otherChainFilePath)

						expectedErr := multierror.NewMultiError("some-command-name")
						expectedErr.Add(errors.New("certificate and chain mismatch: x509: certificate signed by unknown authority"))
						Expect(err).To(Equal(expectedErr))
					})
				})

				Context("if the cert, key and chain are incompatible", func() {
					It("returns multiple errors", func() {
						err := certificateValidator.Validate("some-command-name", certFilePath, otherKeyFilePath, otherChainFilePath)
						expectedErr := multierror.NewMultiError("some-command-name")
						expectedErr.Add(errors.New("tls: private key does not match public key"))
						expectedErr.Add(errors.New("certificate and chain mismatch: x509: certificate signed by unknown authority"))

						Expect(err).To(Equal(expectedErr))
					})
				})
			})

			Context("failure cases", func() {
				Context("when the certificate, key, and chain cannot be read", func() {
					var (
						certFile  string
						keyFile   string
						chainFile string
					)

					BeforeEach(func() {
						createTempFile := func() string {
							file, err := ioutil.TempFile("", "")
							Expect(err).NotTo(HaveOccurred())
							defer file.Close()

							err = os.Chmod(file.Name(), 0100)
							Expect(err).NotTo(HaveOccurred())

							return file.Name()
						}

						certFile = createTempFile()
						keyFile = createTempFile()
						chainFile = createTempFile()
					})

					It("returns an error", func() {
						err := certificateValidator.Validate("some-command-name", certFile, keyFile, chainFile)
						expectedErr := multierror.NewMultiError("some-command-name")
						expectedErr.Add(fmt.Errorf("open %s: permission denied", certFile))
						expectedErr.Add(fmt.Errorf("open %s: permission denied", keyFile))
						expectedErr.Add(fmt.Errorf("open %s: permission denied", chainFile))

						Expect(err).To(Equal(expectedErr))
					})
				})

				Context("when file info cannot be retrieved", func() {
					BeforeEach(func() {
						certs.SetStat(func(string) (os.FileInfo, error) {
							return nil, errors.New("failed to retrieve file info")
						})
					})

					It("returns an error", func() {
						err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, chainFilePath)

						expectedErr := multierror.NewMultiError("some-command-name")
						expectedErr.Add(fmt.Errorf("failed to retrieve file info: %s", certFilePath))
						expectedErr.Add(fmt.Errorf("failed to retrieve file info: %s", keyFilePath))
						expectedErr.Add(fmt.Errorf("failed to retrieve file info: %s", chainFilePath))

						Expect(err).To(Equal(expectedErr))
					})
				})

				Context("when the file cannot be read", func() {
					BeforeEach(func() {
						certs.SetReadAll(func(io.Reader) ([]byte, error) {
							return []byte{}, errors.New("bad read")
						})
					})

					It("returns an error", func() {
						err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, chainFilePath)

						expectedErr := multierror.NewMultiError("some-command-name")
						expectedErr.Add(fmt.Errorf("bad read: %s", certFilePath))
						expectedErr.Add(fmt.Errorf("bad read: %s", keyFilePath))
						expectedErr.Add(fmt.Errorf("bad read: %s", chainFilePath))

						Expect(err).To(Equal(expectedErr))
					})
				})

				Context("when provided files are not valid", func() {
					var file *os.File

					BeforeEach(func() {
						var err error
						file, err = ioutil.TempFile("", "")
						Expect(err).NotTo(HaveOccurred())
						defer file.Close()
					})

					Context("when the private key is not valid rsa", func() {
						BeforeEach(func() {
							err := ioutil.WriteFile(file.Name(), []byte(`
-----BEGIN RSA PRIVATE KEY-----
-----END RSA PRIVATE KEY-----
				`), os.ModePerm)
							Expect(err).NotTo(HaveOccurred())
						})

						It("returns an error", func() {
							err := certificateValidator.Validate("some-command-name", certFilePath, file.Name(), chainFilePath)
							expectedErr := multierror.NewMultiError("some-command-name")
							expectedErr.Add(errors.New("tls: failed to parse private key"))

							Expect(err).To(Equal(expectedErr))
						})
					})

					Context("when the certificate is not valid", func() {
						BeforeEach(func() {
							err := ioutil.WriteFile(file.Name(), []byte(`
-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----
				`), os.ModePerm)
							Expect(err).NotTo(HaveOccurred())
						})

						It("returns an error", func() {
							err := certificateValidator.Validate("some-command-name", file.Name(), keyFilePath, chainFilePath)
							expectedErr := multierror.NewMultiError("some-command-name")
							expectedErr.Add(errors.New("asn1: syntax error: sequence truncated"))

							Expect(err).To(Equal(expectedErr))
						})
					})

					Context("when the chain is not valid", func() {
						BeforeEach(func() {
							err := ioutil.WriteFile(file.Name(), []byte(`
-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----
				`), os.ModePerm)
							Expect(err).NotTo(HaveOccurred())
						})

						It("returns an error", func() {
							err := certificateValidator.Validate("some-command-name", certFilePath, keyFilePath, file.Name())
							expectedErr := multierror.NewMultiError("some-command-name")
							expectedErr.Add(errors.New("failed to parse chain"))

							Expect(err).To(Equal(expectedErr))
						})
					})
				})
			})
		})

		Context("when using a PKCS#8 key", func() {
			Context("when cert and key are valid", func() {
				It("does not return an error", func() {
					err := certificateValidator.Validate("some-command-name", "fixtures/pkcs8.crt", "fixtures/pkcs8.key", "")

					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
