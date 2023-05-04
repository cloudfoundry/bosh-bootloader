package bosh_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CredhubGetter", func() {
	var (
		credhubGetter bosh.CredhubGetter
		stateStore    *fakes.StateStore
		fileIO        *fakes.FileIO
	)

	BeforeEach(func() {
		stateStore = &fakes.StateStore{}
		stateStore.GetVarsDirCall.Returns.Directory = "fake-vars-dir"
		fileIO = &fakes.FileIO{}
		credhubGetter = bosh.NewCredhubGetter(stateStore, fileIO)
	})

	Context("reading from the vars store", func() {
		BeforeEach(func() {
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
credhub_admin_client_secret: some-credhub-password`
			fileIO.ReadFileCall.Returns.Contents = []byte(varsStoreContents)
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
						fileIO.ReadFileCall.Returns.Contents = []byte("invalid yaml")
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
						fileIO.ReadFileCall.Returns.Error = errors.New("passionfruit")
					})

					It("returns an error", func() {
						_, err := credhubGetter.GetCerts()
						Expect(err).To(MatchError(ContainSubstring("Read director-vars-store.yml file: ")))
						Expect(err).To(MatchError(ContainSubstring("passionfruit")))
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
						fileIO.ReadFileCall.Returns.Contents = []byte("invalid yaml")
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
						fileIO.ReadFileCall.Returns.Error = errors.New("quamquat")
					})

					It("returns an error", func() {
						_, err := credhubGetter.GetCerts()
						Expect(err).To(MatchError(ContainSubstring("Read director-vars-store.yml file: ")))
						Expect(err).To(MatchError(ContainSubstring("quamquat")))
					})
				})
			})
		})
	})

	Describe("GetServer", func() {
		BeforeEach(func() {
			varsFileContents := `---
internal_ip: some-internal-ip`
			fileIO.ReadFileCall.Returns.Contents = []byte(varsFileContents)
		})

		It("returns the credhub server url", func() {
			server, err := credhubGetter.GetServer()
			Expect(err).NotTo(HaveOccurred())
			Expect(server).To(Equal("https://some-internal-ip:8844"))
		})

		Context("failure cases", func() {
			Context("when the vars file cannot be unmarshaled", func() {
				BeforeEach(func() {
					fileIO.ReadFileCall.Returns.Contents = []byte("invalid yaml")
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
					fileIO.ReadFileCall.Returns.Error = errors.New("quamquat")
				})

				It("returns an error", func() {
					_, err := credhubGetter.GetServer()
					Expect(err).To(MatchError(ContainSubstring("Read director-vars-file.yml file: ")))
					Expect(err).To(MatchError(ContainSubstring("quamquat")))
				})
			})
		})
	})
})
