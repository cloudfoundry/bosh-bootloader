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

var _ = Describe("LB args handler", func() {
	var (
		handler              commands.LBArgsHandler
		certificateValidator *fakes.CertificateValidator
	)

	BeforeEach(func() {
		certificateValidator = &fakes.CertificateValidator{}
		handler = commands.NewLBArgsHandler(certificateValidator)
	})

	Describe("GetLBState", func() {
		BeforeEach(func() {
			certificateValidator.ReadAndValidateCall.Returns.CertData = certs.CertData{
				Key:   []byte("some-key"),
				Cert:  []byte("some-cert"),
				Chain: []byte("some-chain"),
			}
		})

		It("returns a storage.LB object", func() {
			lbState, err := handler.GetLBState("aws", commands.CreateLBsConfig{
				LBType:    "cf",
				CertPath:  "/path/to/cert",
				KeyPath:   "/path/to/key",
				ChainPath: "/path/to/chain",
				Domain:    "something.io",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(lbState.Type).To(Equal("cf"))
			Expect(lbState.Cert).To(Equal("some-cert"))
			Expect(lbState.Key).To(Equal("some-key"))
			Expect(lbState.Chain).To(Equal("some-chain"))
			Expect(lbState.Domain).To(Equal("something.io"))

			Expect(certificateValidator.ReadAndValidateCall.CallCount).To(Equal(1))
			Expect(certificateValidator.ReadAndValidateCall.Receives.CertificatePath).To(Equal("/path/to/cert"))
			Expect(certificateValidator.ReadAndValidateCall.Receives.KeyPath).To(Equal("/path/to/key"))
			Expect(certificateValidator.ReadAndValidateCall.Receives.ChainPath).To(Equal("/path/to/chain"))
		})

		Context("when iaas is gcp and lb type is concourse", func() {
			It("does not call certificateValidator", func() {
				lbState, err := handler.GetLBState("gcp", commands.CreateLBsConfig{
					LBType: "concourse",
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(lbState.Type).To(Equal("concourse"))
				Expect(lbState.Cert).To(Equal(""))
				Expect(lbState.Key).To(Equal(""))
				Expect(lbState.Chain).To(Equal(""))
				Expect(lbState.Domain).To(Equal(""))
				Expect(certificateValidator.ReadAndValidateCall.CallCount).To(Equal(0))
			})
		})

		Context("failure cases", func() {
			Context("if there is no lb type", func() {
				It("returns an error", func() {
					_, err := handler.GetLBState("", commands.CreateLBsConfig{})
					Expect(err).To(MatchError("--type is required"))
				})
			})

			Context("when certificate validator fails for cert and key", func() {
				It("returns an error", func() {
					certificateValidator.ReadAndValidateCall.Returns.Error = errors.New("failed to validate")
					_, err := handler.GetLBState("aws", commands.CreateLBsConfig{
						LBType:    "concourse",
						CertPath:  "/path/to/cert",
						KeyPath:   "/path/to/key",
						ChainPath: "/path/to/chain",
					})

					Expect(err).To(MatchError("Validate certificate: failed to validate"))
				})
			})

			Context("when lb type is concourse and domain flag is supplied", func() {
				It("returns an error", func() {
					_, err := handler.GetLBState("gcp", commands.CreateLBsConfig{
						LBType: "concourse",
						Domain: "something.io",
					})
					Expect(err).To(MatchError("--domain is not implemented for concourse load balancers. Remove the --domain flag and try again."))
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
				Chain:  "new-chain",
				Domain: "new-domain",
			}
			old = storage.LB{
				Type:   "old-type",
				Cert:   "old-cert",
				Key:    "old-key",
				Chain:  "old-chain",
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
