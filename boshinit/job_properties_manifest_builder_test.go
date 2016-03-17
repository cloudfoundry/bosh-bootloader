package boshinit_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
)

var _ = Describe("JobPropertiesManifestBuilder", func() {
	var (
		jobPropertiesManifestBuilder boshinit.JobPropertiesManifestBuilder
		natsUsername                 string
		postgresUsername             string
		registryUsername             string
		blobstoreDirectorUsername    string
		blobstoreAgentUsername       string
		hmUsername                   string
		natsPassword                 string
		redisPassword                string
		postgresPassword             string
		registryPassword             string
		blobstoreDirectorPassword    string
		blobstoreAgentPassword       string
		hmPassword                   string
	)

	BeforeEach(func() {
		natsUsername = "random-nats-username"
		postgresUsername = "random-postgres-username"
		registryUsername = "random-registry-username"
		blobstoreDirectorUsername = "random-blobstore-director-username"
		blobstoreAgentUsername = "random-blobstore-agent-username"
		hmUsername = "random-hm-username"
		natsPassword = "random-nats-password"
		redisPassword = "random-redis-password"
		postgresPassword = "random-postgres-password"
		registryPassword = "random-registry-password"
		blobstoreDirectorPassword = "random-blobstore-director-password"
		blobstoreAgentPassword = "random-blobstore-agent-password"
		hmPassword = "random-hm-password"
		jobPropertiesManifestBuilder = boshinit.NewJobPropertiesManifestBuilder(
			natsUsername,
			postgresUsername,
			registryUsername,
			blobstoreDirectorUsername,
			blobstoreAgentUsername,
			hmUsername,
			natsPassword,
			redisPassword,
			postgresPassword,
			registryPassword,
			blobstoreDirectorPassword,
			blobstoreAgentPassword,
			hmPassword,
		)
	})

	Describe("NATS", func() {
		It("returns job properties for NATS", func() {
			nats := jobPropertiesManifestBuilder.NATS()
			Expect(nats).To(Equal(
				boshinit.NATSJobProperties{
					Address:  "127.0.0.1",
					User:     natsUsername,
					Password: natsPassword,
				}))
		})
	})

	Describe("Redis", func() {
		It("returns job properties for Redis", func() {
			redis := jobPropertiesManifestBuilder.Redis()
			Expect(redis).To(Equal(
				boshinit.RedisJobProperties{
					ListenAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Password:      redisPassword,
				}))
		})
	})

	Describe("Postgres", func() {
		It("returns job properties for Postgres", func() {
			postgres := jobPropertiesManifestBuilder.Postgres()
			Expect(postgres).To(Equal(boshinit.PostgresProperties{
				ListenAddress: "127.0.0.1",
				Host:          "127.0.0.1",
				User:          postgresUsername,
				Password:      postgresPassword,
				Database:      "bosh",
				Adapter:       "postgres",
			}))
		})
	})

	Describe("Registry", func() {
		It("returns job properties for Registry", func() {
			registry := jobPropertiesManifestBuilder.Registry()
			Expect(registry).To(Equal(boshinit.RegistryJobProperties{
				Address:  "10.0.0.6",
				Host:     "10.0.0.6",
				Username: registryUsername,
				Password: registryPassword,
				Port:     25777,
				DB: boshinit.PostgresProperties{
					ListenAddress: "127.0.0.1",
					Host:          "127.0.0.1",
					User:          postgresUsername,
					Password:      postgresPassword,
					Database:      "bosh",
					Adapter:       "postgres",
				},
				HTTP: boshinit.HTTPProperties{
					User:     registryUsername,
					Password: registryPassword,
					Port:     25777,
				},
			}))
		})
	})

	Describe("Blobstore", func() {
		It("returns job properties for Blobstore", func() {
			blobstore := jobPropertiesManifestBuilder.Blobstore()
			Expect(blobstore).To(Equal(boshinit.BlobstoreJobProperties{
				Address:  "10.0.0.6",
				Port:     25250,
				Provider: "dav",
				Director: boshinit.Credentials{
					User:     blobstoreDirectorUsername,
					Password: blobstoreDirectorPassword,
				},
				Agent: boshinit.Credentials{
					User:     blobstoreAgentUsername,
					Password: blobstoreAgentPassword,
				},
			}))
		})
	})

	Describe("Director", func() {
		It("returns job properties for Director", func() {
			director := jobPropertiesManifestBuilder.Director(boshinit.ManifestProperties{
				DirectorUsername: "bosh-username",
				DirectorPassword: "bosh-password",
				SSLKeyPair: ssl.KeyPair{
					Certificate: []byte("some-ssl-cert"),
					PrivateKey:  []byte("some-ssl-key"),
				},
			})
			Expect(director).To(Equal(boshinit.DirectorJobProperties{
				Address:    "127.0.0.1",
				Name:       "my-bosh",
				CPIJob:     "aws_cpi",
				MaxThreads: 10,
				DB: boshinit.PostgresProperties{
					ListenAddress: "127.0.0.1",
					Host:          "127.0.0.1",
					User:          postgresUsername,
					Password:      postgresPassword,
					Database:      "bosh",
					Adapter:       "postgres",
				},
				UserManagement: boshinit.UserManagementProperties{
					Provider: "local",
					Local: boshinit.LocalProperties{
						Users: []boshinit.UserProperties{
							{
								Name:     "bosh-username",
								Password: "bosh-password",
							},
							{
								Name:     hmUsername,
								Password: hmPassword,
							},
						},
					},
				},
				SSL: boshinit.SSLProperties{
					Cert: "some-ssl-cert",
					Key:  "some-ssl-key",
				},
			}))
		})
	})

	Describe("HM", func() {
		It("returns job properties for HM", func() {
			hm := jobPropertiesManifestBuilder.HM()
			Expect(hm).To(Equal(boshinit.HMJobProperties{
				DirectorAccount: boshinit.Credentials{
					User:     hmUsername,
					Password: hmPassword,
				},
				ResurrectorEnabled: true,
			}))
		})
	})

	Describe("Agent", func() {
		It("returns job properties for Agent", func() {
			agent := jobPropertiesManifestBuilder.Agent()
			Expect(agent).To(Equal(boshinit.AgentProperties{
				MBus: "nats://random-nats-username:random-nats-password@10.0.0.6:4222",
			}))
		})
	})
})
