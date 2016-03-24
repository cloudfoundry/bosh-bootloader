package manifests_test

import (
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit/manifests"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InternalCredentials", func() {
	var (
		inputCredentials map[string]string
	)

	BeforeEach(func() {
		inputCredentials = map[string]string{
			"mbusUsername":              "some-mbus-username",
			"natsUsername":              "some-nats-username",
			"postgresUsername":          "some-postgres-username",
			"registryUsername":          "some-registry-username",
			"blobstoreDirectorUsername": "some-blobstore-director-username",
			"blobstoreAgentUsername":    "some-blobstore-agent-username",
			"hmUsername":                "some-hm-username",
			"mbusPassword":              "some-mbus-password",
			"natsPassword":              "some-nats-password",
			"redisPassword":             "some-redis-password",
			"postgresPassword":          "some-postgres-password",
			"registryPassword":          "some-registry-password",
			"blobstoreDirectorPassword": "some-blobstore-director-password",
			"blobstoreAgentPassword":    "some-blobstore-agent-password",
			"hmPassword":                "some-hm-password",
		}
	})

	Describe("NewInternalCredentials", func() {
		It("creates a new InternalCredentials from a map[string]string", func() {
			outputCredentials := manifests.NewInternalCredentials(inputCredentials)

			Expect(outputCredentials).To(Equal(manifests.InternalCredentials{
				MBusUsername:              "some-mbus-username",
				NatsUsername:              "some-nats-username",
				PostgresUsername:          "some-postgres-username",
				RegistryUsername:          "some-registry-username",
				BlobstoreDirectorUsername: "some-blobstore-director-username",
				BlobstoreAgentUsername:    "some-blobstore-agent-username",
				HMUsername:                "some-hm-username",
				MBusPassword:              "some-mbus-password",
				NatsPassword:              "some-nats-password",
				RedisPassword:             "some-redis-password",
				PostgresPassword:          "some-postgres-password",
				RegistryPassword:          "some-registry-password",
				BlobstoreDirectorPassword: "some-blobstore-director-password",
				BlobstoreAgentPassword:    "some-blobstore-agent-password",
				HMPassword:                "some-hm-password",
			}))
		})
	})

	Describe("ToMap", func() {
		It("returns a map representation of the credentials", func() {
			outputCredentials := manifests.NewInternalCredentials(inputCredentials)
			Expect(outputCredentials.ToMap()).To(Equal(inputCredentials))
		})
	})
})
