package boshinit_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
)

var _ = Describe("SharedPropertiesManifestBuilder", func() {
	var sharedPropertiesManifestBuilder boshinit.SharedPropertiesManifestBuilder

	BeforeEach(func() {
		sharedPropertiesManifestBuilder = boshinit.NewSharedPropertiesManifestBuilder()
	})

	Describe("Postgres", func() {
		It("returns job properties for Postgres", func() {
			postgres := sharedPropertiesManifestBuilder.Postgres()
			Expect(postgres).To(Equal(boshinit.PostgresProperties{
				ListenAddress: "127.0.0.1",
				Host:          "127.0.0.1",
				User:          "postgres",
				Password:      "postgres-password",
				Database:      "bosh",
				Adapter:       "postgres",
			}))
		})
	})

	Describe("AWS", func() {
		It("returns job properties for AWS", func() {
			aws := sharedPropertiesManifestBuilder.AWS()
			Expect(aws).To(Equal(boshinit.AWSProperties{
				AccessKeyId:           "ACCESS-KEY-ID",
				SecretAccessKey:       "SECRET-ACCESS-KEY",
				DefaultKeyName:        "bosh",
				DefaultSecurityGroups: []string{"bosh"},
				Region:                "REGION",
			}))
		})
	})

	Describe("NTP", func() {
		It("returns job properties for NTP", func() {
			ntp := sharedPropertiesManifestBuilder.NTP()
			Expect(ntp).To(ConsistOf(
				[]string{"0.pool.ntp.org", "1.pool.ntp.org"},
			))
		})
	})
})
