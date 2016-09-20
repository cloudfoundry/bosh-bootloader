package bosh_test

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CIDRBlock", func() {
	var (
		cidrBlock bosh.CIDRBlock
	)

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

	Describe("GetLastIP", func() {
		It("returns the first ip of the cidr block", func() {
			ip := cidrBlock.GetLastIP()
			Expect(ip.String()).To(Equal("10.0.31.255"))
		})
	})

	Describe("ParseCIDRBlock", func() {
		Context("failure cases", func() {
			It("returns an error when input string is not a valid CIDR block", func() {
				_, err := bosh.ParseCIDRBlock("whatever")
				Expect(err).To(MatchError(ContainSubstring("cannot parse CIDR block")))
			})

			It("returns an error when input string contains an invalid ip", func() {
				_, err := bosh.ParseCIDRBlock("not-an-ip/20")
				Expect(err).To(MatchError(ContainSubstring("not a valid ip address")))
			})

			It("returns an error when input string contains mask bits which are not an integer", func() {
				_, err := bosh.ParseCIDRBlock("0.0.0.0/not-mask-bits")
				Expect(err).To(MatchError(ContainSubstring("invalid syntax")))
			})

			It("returns an error when input string contains mask bits which are out of range", func() {
				_, err := bosh.ParseCIDRBlock("0.0.0.0/243")
				Expect(err).To(MatchError(ContainSubstring("mask bits out of range")))
			})
		})
	})
})
