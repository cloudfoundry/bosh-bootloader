package bosh_test

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IP", func() {
	Context("v4", func() {
		Describe("ParseIP", func() {
			It("returns an IP object that represents IP from string", func() {
				ip, err := bosh.ParseIP("10.0.16.255")
				Expect(err).NotTo(HaveOccurred())
				Expect(ip.String()).To(Equal("10.0.16.255"))
			})

			Context("failure cases", func() {
				It("returns an error if it cannot parse ip", func() {
					_, err := bosh.ParseIP("not valid")
					Expect(err).To(MatchError(ContainSubstring("unable to parse IP")))
				})

				It("returns an error if ip parts are not digits", func() {
					_, err := bosh.ParseIP("x.x.x.x")
					Expect(err).To(MatchError(ContainSubstring("unexpected character")))
				})

				It("returns an error if ip parts are out of the allowed range", func() {
					_, err := bosh.ParseIP("999.999.999.999")
					Expect(err).To(MatchError(ContainSubstring("IPv4 field has value >255")))
				})

				It("returns an error if ip has too many parts", func() {
					_, err := bosh.ParseIP("1.1.1.1.1.1.1")
					Expect(err).To(MatchError(ContainSubstring("IPv4 address too long")))
				})
			})
		})

		Describe("Add", func() {
			It("returns an IP object that represents IP offsetted by 1", func() {
				ip, err := bosh.ParseIP("10.0.16.1")
				ip = ip.Add(1)
				Expect(err).NotTo(HaveOccurred())
				Expect(ip.String()).To(Equal("10.0.16.2"))
			})
		})

		Describe("Subtract", func() {
			It("returns an IP object that represents IP offsetted by -1", func() {
				ip, err := bosh.ParseIP("10.0.16.2")
				ip = ip.Subtract(1)
				Expect(err).NotTo(HaveOccurred())
				Expect(ip.String()).To(Equal("10.0.16.1"))
			})
		})

		Describe("String", func() {
			It("returns a string representation of IP object", func() {
				ip, err := bosh.ParseIP("10.0.16.1")
				Expect(err).NotTo(HaveOccurred())
				Expect(ip.String()).To(Equal("10.0.16.1"))
			})
		})
	})
	Context("v6", func() {
		Describe("ParseIP", func() {
			It("returns an IP object that represents IP from string", func() {
				ip, err := bosh.ParseIP("2001:db8:cf:0:0:0:ffff:1337")
				Expect(err).NotTo(HaveOccurred())
				Expect(ip.String()).To(Equal("2001:db8:cf::ffff:1337"))
			})

			Context("failure cases", func() {
				It("returns an error if it cannot parse ip", func() {
					_, err := bosh.ParseIP("2001:db8:cf::not valid")
					Expect(err).To(MatchError(ContainSubstring("each colon-separated field must have at least one digit")))
				})

				It("returns an error if ip parts are not digits", func() {
					_, err := bosh.ParseIP("2001:db8:cf:x:x:x:x")
					Expect(err).To(MatchError(ContainSubstring("each colon-separated field must have at least one digit")))
				})

				It("returns an error if ip parts are out of the allowed range", func() {
					_, err := bosh.ParseIP("2001:db8:cf:G::")
					Expect(err).To(MatchError(ContainSubstring("each colon-separated field must have at least one digit")))
				})

				It("returns an error if ip has too many parts", func() {
					_, err := bosh.ParseIP("2001:db8:cf:0:0:0:ffff:ffff:ffff")
					Expect(err).To(MatchError(ContainSubstring("trailing garbage after address")))
				})
			})
		})

		Describe("Add", func() {
			It("returns an IP object that represents IP offsetted by 1", func() {
				ip, err := bosh.ParseIP("2001:db8:cf::")
				ip = ip.Add(1)
				Expect(err).NotTo(HaveOccurred())
				Expect(ip.String()).To(Equal("2001:db8:cf::1"))
			})
		})

		Describe("Subtract", func() {
			It("returns an IP object that represents IP offsetted by -1", func() {
				ip, err := bosh.ParseIP("2001:db8:cf::2")
				ip = ip.Subtract(1)
				Expect(err).NotTo(HaveOccurred())
				Expect(ip.String()).To(Equal("2001:db8:cf::1"))
			})
		})

		Describe("String", func() {
			It("returns a string representation of IP object", func() {
				ip, err := bosh.ParseIP("2001:db8:cf::ffff:ffff:ffff")
				Expect(err).NotTo(HaveOccurred())
				Expect(ip.String()).To(Equal("2001:db8:cf::ffff:ffff:ffff"))
			})
		})
	})
})
