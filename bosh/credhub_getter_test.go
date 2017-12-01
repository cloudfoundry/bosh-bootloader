package bosh_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CredhubGetter", func() {
	var (
		credhubGetter bosh.CredhubGetter
		stateStore    *fakes.StateStore
		varsDir       string
		varsFilePath  string
		varsStorePath string
	)

	BeforeEach(func() {
		var err error
		varsDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
		stateStore = &fakes.StateStore{}
		stateStore.GetVarsDirCall.Returns.Directory = varsDir
		credhubGetter = bosh.NewCredhubGetter(stateStore)
	})

	Context("reading from the vars store", func() {
		BeforeEach(func() {
			varsStorePath = filepath.Join(varsDir, "director-vars-store.yml")
			varsStoreContents := `---
credhub_ca:
  certificate: |
    -----BEGIN CERTIFICATE-----
    some-credhub-cert
    -----END CERTIFICATE-----
uaa_ssl:
  certificate: |
    -----BEGIN CERTIFICATE-----
    some-uaa-cert
    -----END CERTIFICATE-----
credhub_cli_password: some-credhub-password`
			err := ioutil.WriteFile(varsStorePath, []byte(varsStoreContents), storage.ScriptMode)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("GetCerts", func() {
			It("returns the credhub cert and uaa cert", func() {
				certs, err := credhubGetter.GetCerts()
				Expect(err).NotTo(HaveOccurred())
				Expect(certs).To(Equal(`-----BEGIN CERTIFICATE-----
some-credhub-cert
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
some-uaa-cert
-----END CERTIFICATE-----
`))
			})

			Context("failure cases", func() {
				Context("when the vars store cannot be unmarshaled", func() {
					BeforeEach(func() {
						err := ioutil.WriteFile(varsStorePath, []byte("invalid yaml"), storage.ScriptMode)
						Expect(err).NotTo(HaveOccurred())
					})

					It("returns an error", func() {
						_, err := credhubGetter.GetCerts()
						Expect(err).To(MatchError(ContainSubstring("line 1: cannot unmarshal !!str `invalid...`")))
					})
				})

				Context("when the state store fails to get the vars dir", func() {
					BeforeEach(func() {
						stateStore.GetVarsDirCall.Returns.Error = errors.New("tangelo")
					})

					It("returns an error", func() {
						_, err := credhubGetter.GetCerts()
						Expect(err).To(MatchError("Get vars directory: tangelo"))
					})
				})

				Context("when the vars store can't be read", func() {
					BeforeEach(func() {
						err := os.Remove(varsStorePath)
						Expect(err).NotTo(HaveOccurred())
					})

					It("returns an error", func() {
						_, err := credhubGetter.GetCerts()
						Expect(err).To(MatchError(ContainSubstring("Read director-vars-store.yml file: ")))
						Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
					})
				})
			})
		})

		Describe("GetPassword", func() {
			It("returns the credhub password", func() {
				certs, err := credhubGetter.GetPassword()
				Expect(err).NotTo(HaveOccurred())
				Expect(certs).To(Equal("some-credhub-password"))
			})

			Context("failure cases", func() {
				Context("when the vars store cannot be unmarshaled", func() {
					BeforeEach(func() {
						err := ioutil.WriteFile(varsStorePath, []byte("invalid yaml"), storage.ScriptMode)
						Expect(err).NotTo(HaveOccurred())
					})

					It("returns an error", func() {
						_, err := credhubGetter.GetCerts()
						Expect(err).To(MatchError(ContainSubstring("line 1: cannot unmarshal !!str `invalid...`")))
					})
				})

				Context("when the state store fails to get the vars dir", func() {
					BeforeEach(func() {
						stateStore.GetVarsDirCall.Returns.Error = errors.New("tangelo")
					})

					It("returns an error", func() {
						_, err := credhubGetter.GetCerts()
						Expect(err).To(MatchError("Get vars directory: tangelo"))
					})
				})

				Context("when the vars store can't be read", func() {
					BeforeEach(func() {
						err := os.Remove(varsStorePath)
						Expect(err).NotTo(HaveOccurred())
					})

					It("returns an error", func() {
						_, err := credhubGetter.GetCerts()
						Expect(err).To(MatchError(ContainSubstring("Read director-vars-store.yml file: ")))
						Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
					})
				})
			})
		})
	})

	Describe("GetServer", func() {
		BeforeEach(func() {
			varsFilePath = filepath.Join(varsDir, "director-vars-file.yml")
			varsFileContents := `---
internal_ip: some-internal-ip`
			err := ioutil.WriteFile(varsFilePath, []byte(varsFileContents), storage.ScriptMode)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the credhub server url", func() {
			server, err := credhubGetter.GetServer()
			Expect(err).NotTo(HaveOccurred())
			Expect(server).To(Equal("https://some-internal-ip:8844"))
		})

		Context("failure cases", func() {
			Context("when the vars file cannot be unmarshaled", func() {
				BeforeEach(func() {
					err := ioutil.WriteFile(varsFilePath, []byte("invalid yaml"), storage.ScriptMode)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					_, err := credhubGetter.GetServer()
					Expect(err).To(MatchError(ContainSubstring("line 1: cannot unmarshal !!str `invalid...`")))
				})
			})

			Context("when the state store fails to get the vars dir", func() {
				BeforeEach(func() {
					stateStore.GetVarsDirCall.Returns.Error = errors.New("tangelo")
				})

				It("returns an error", func() {
					_, err := credhubGetter.GetServer()
					Expect(err).To(MatchError("Get vars directory: tangelo"))
				})
			})

			Context("when the vars file can't be read", func() {
				BeforeEach(func() {
					err := os.Remove(varsFilePath)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					_, err := credhubGetter.GetServer()
					Expect(err).To(MatchError(ContainSubstring("Read director-vars-file.yml file: ")))
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})
		})
	})
})
