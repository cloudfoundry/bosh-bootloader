package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/certs"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LBArgsHandler", func() {
	var (
		certificateValidator *fakes.CertificateValidator
		handler              commands.LBArgsHandler
	)

	BeforeEach(func() {
		certificateValidator = &fakes.CertificateValidator{}
		handler = commands.NewLBArgsHandler(certificateValidator)
	})

	Describe("GetLBState", func() {
		BeforeEach(func() {
			certData := certs.CertData{
				Key:  []byte("some-key"),
				Cert: []byte("some-cert"),
			}
			certificateValidator.ReadAndValidateCall.Returns.CertData = certData
			certificateValidator.ReadCall.Returns.CertData = certData
		})

		It("returns a storage.LB object", func() {
			lbState, err := handler.GetLBState("aws", commands.LBArgs{
				LBType:   "cf",
				CertPath: "/path/to/cert",
				KeyPath:  "/path/to/key",
				Domain:   "something.io",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(lbState.Type).To(Equal("cf"))
			Expect(lbState.Cert).To(Equal("some-cert"))
			Expect(lbState.Key).To(Equal("some-key"))
			Expect(lbState.Domain).To(Equal("something.io"))

			Expect(certificateValidator.ReadAndValidateCall.CallCount).To(Equal(1))
			Expect(certificateValidator.ReadAndValidateCall.Receives.CertificatePath).To(Equal("/path/to/cert"))
			Expect(certificateValidator.ReadAndValidateCall.Receives.KeyPath).To(Equal("/path/to/key"))
		})

		Context("when lb type is concourse", func() {
			Context("on gcp", func() {
				It("does not call certificateValidator", func() {
					lbState, err := handler.GetLBState("gcp", commands.LBArgs{
						LBType: "concourse",
					})
					Expect(err).NotTo(HaveOccurred())
					Expect(lbState.Type).To(Equal("concourse"))
					Expect(lbState.Cert).To(Equal(""))
					Expect(lbState.Key).To(Equal(""))
					Expect(lbState.Domain).To(Equal(""))
					Expect(certificateValidator.ReadAndValidateCall.CallCount).To(Equal(0))
				})
			})
			Context("on aws", func() {
				It("does not call certificateValidator", func() {
					lbState, err := handler.GetLBState("aws", commands.LBArgs{LBType: "concourse"})
					Expect(err).NotTo(HaveOccurred())

					Expect(lbState.Type).To(Equal("concourse"))
					Expect(certificateValidator.ReadAndValidateCall.CallCount).To(Equal(0))
				})
			})
		})

		Context("when iaas is azure and lb type is cf", func() {
			BeforeEach(func() {
				certDataPKCS12 := certs.CertData{
					Cert: []byte("some-cert"),
					Key:  []byte("some-password"),
				}
				certificateValidator.ReadAndValidatePKCS12Call.Returns.CertData = certDataPKCS12
				certificateValidator.ReadPKCS12Call.Returns.CertData = certDataPKCS12
			})

			It("it reads the certificate and validates the password", func() {
				lbState, err := handler.GetLBState("azure", commands.LBArgs{LBType: "cf"})
				Expect(err).NotTo(HaveOccurred())

				Expect(lbState.Type).To(Equal("cf"))
				Expect(lbState.Cert).To(Equal("c29tZS1jZXJ0"))
				Expect(lbState.Key).To(Equal("some-password"))
				Expect(certificateValidator.ReadAndValidatePKCS12Call.CallCount).To(Equal(1))
			})
		})

		Context("when empty config is passed in", func() {
			It("does not call certificateValidator", func() {
				_, err := handler.GetLBState("", commands.LBArgs{})
				Expect(err).NotTo(HaveOccurred())
				Expect(certificateValidator.ReadAndValidateCall.CallCount).To(Equal(0))
			})
		})

		Context("failure cases", func() {
			Context("when certificate validator fails for aws/gcp", func() {
				It("returns an error", func() {
					certificateValidator.ReadAndValidateCall.Returns.Error = errors.New("failed to validate")

					_, err := handler.GetLBState("aws", commands.LBArgs{LBType: "cf"})
					Expect(err).To(MatchError("Validate certificate: failed to validate"))
				})
			})

			Context("when certificate validator fails for azure", func() {
				It("returns an error", func() {
					certificateValidator.ReadAndValidatePKCS12Call.Returns.Error = errors.New("failed to validate")

					_, err := handler.GetLBState("azure", commands.LBArgs{LBType: "cf"})
					Expect(err).To(MatchError("Validate certificate: failed to validate"))
				})
			})

			Context("when lb type is concourse and domain flag is supplied", func() {
				It("returns an error", func() {
					_, err := handler.GetLBState("gcp", commands.LBArgs{
						LBType: "concourse",
						Domain: "something.io",
					})
					Expect(err).To(MatchError("domain is not implemented for concourse load balancers. Remove the --lb-domain flag and try again."))
				})
			})
		})
	})

	Describe("Merge", func() {
		var new storage.LB
		var old storage.LB

		BeforeEach(func() {
			new = storage.LB{
				Type:   "new-type",
				Cert:   "new-cert",
				Key:    "new-key",
				Domain: "new-domain",
			}
			old = storage.LB{
				Type:   "old-type",
				Cert:   "old-cert",
				Key:    "old-key",
				Domain: "old-domain",
			}
		})

		Context("when the old state is empty", func() {
			It("returns the new state", func() {
				merged := handler.Merge(new, storage.LB{})
				Expect(merged).To(Equal(new))
			})
		})

		Context("when the new state has all fields populated", func() {
			It("returns the new state", func() {
				merged := handler.Merge(new, old)
				Expect(merged).To(Equal(new))
			})
		})

		Context("when the new state is empty", func() {
			It("keeps the old domain and type", func() {
				merged := handler.Merge(storage.LB{}, old)
				Expect(merged).To(Equal(storage.LB{
					Type:   "old-type",
					Domain: "old-domain",
				}))
			})
		})
	})
})
