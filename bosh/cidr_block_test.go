package bosh_test

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CIDRBlock", func() {
	var (
		cidrBlock bosh.CIDRBlock
	)
	Context("v4", func() {
		BeforeEach(func() {
			var err error
			cidrBlock, err = bosh.ParseCIDRBlock("10.0.16.0/20")
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("GetFirstIP", func() {
			It("returns the first ip of the cidr block", func() {
				ip := cidrBlock.GetFirstIP()
				Expect(ip.String()).To(Equal("10.0.16.0"))
			})
		})

		Describe("GetNthIP", func() {
			It("returns the nth ip of the cidr block", func() {
				ip := cidrBlock.GetNthIP(6)
				Expect(ip.String()).To(Equal("10.0.16.6"))
			})
		})

		Describe("GetLastIP", func() {
			It("returns the first ip of the cidr block", func() {
				ip := cidrBlock.GetLastIP()
				Expect(ip.String()).To(Equal("10.0.31.255"))
			})
		})

		Describe("ParseCIDRBlock", func() {
			Context("failure cases", func() {
				Context("when input string is not a valid CIDR block", func() {
					It("returns an error", func() {
						_, err := bosh.ParseCIDRBlock("whatever")
						Expect(err).To(MatchError(ContainSubstring("no '/'")))
					})
				})

				Context("when input string contains an invalid ip", func() {
					It("returns an error", func() {
						_, err := bosh.ParseCIDRBlock("not-an-ip/20")
						Expect(err).To(MatchError(ContainSubstring("unable to parse IP")))
					})
				})

				Context("when input string contains mask bits which are not an integer", func() {
					It("returns an error", func() {
						_, err := bosh.ParseCIDRBlock("0.0.0.0/not-mask-bits")
						Expect(err).To(MatchError(ContainSubstring(`bad bits after slash: "not-mask-bits"`)))
					})
				})

				Context("when input string contains mask bits which are out of range", func() {
					It("returns an error", func() {
						_, err := bosh.ParseCIDRBlock("0.0.0.0/243")
						Expect(err).To(MatchError(ContainSubstring("prefix length out of range")))
					})
				})
			})
		})
	})
	Context("v6", func() {
		BeforeEach(func() {
			var err error
			cidrBlock, err = bosh.ParseCIDRBlock("2001:db8:cf::/80")
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("GetFirstIP", func() {
			It("returns the first ip of the cidr block", func() {
				ip := cidrBlock.GetFirstIP()
				Expect(ip.String()).To(Equal("2001:db8:cf::"))
			})
		})

		Describe("GetNthIP", func() {
			It("returns the nth ip of the cidr block", func() {
				ip := cidrBlock.GetNthIP(6)
				Expect(ip.String()).To(Equal("2001:db8:cf::6"))
			})
		})

		Describe("GetLastIP", func() {
			It("returns the first ip of the cidr block", func() {
				ip := cidrBlock.GetLastIP()
				Expect(ip.String()).To(Equal("2001:db8:cf::ffff:ffff:ffff"))
			})
		})

		Describe("ParseCIDRBlock", func() {
			Context("failure cases", func() {
				Context("when input string is not a valid CIDR block", func() {
					It("returns an error", func() {
						_, err := bosh.ParseCIDRBlock("whatever")
						Expect(err).To(MatchError(ContainSubstring("no '/'")))
					})
				})

				Context("when input string contains an invalid ip", func() {
					It("returns an error", func() {
						_, err := bosh.ParseCIDRBlock("not-an-ip/96")
						Expect(err).To(MatchError(ContainSubstring("unable to parse IP")))
					})
				})

				Context("when input string contains mask bits which are not an integer", func() {
					It("returns an error", func() {
						_, err := bosh.ParseCIDRBlock("2001:db8:cf::/not-mask-bits")
						Expect(err).To(MatchError(ContainSubstring(`bad bits after slash: "not-mask-bits"`)))
					})
				})

				Context("when input string contains mask bits which are out of range", func() {
					It("returns an error", func() {
						_, err := bosh.ParseCIDRBlock("2001:db8:cf::/243")
						Expect(err).To(MatchError(ContainSubstring("prefix length out of range")))
					})
				})
			})
		})
	})
})
