package manifests_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit/manifests"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
)

var _ = Describe("JobPropertiesManifestBuilder", func() {
	var (
		jobPropertiesManifestBuilder manifests.JobPropertiesManifestBuilder
		natsUsername                 string
		postgresUsername             string
		registryUsername             string
		blobstoreDirectorUsername    string
		blobstoreAgentUsername       string
		hmUsername                   string
		natsPassword                 string
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
		postgresPassword = "random-postgres-password"
		registryPassword = "random-registry-password"
		blobstoreDirectorPassword = "random-blobstore-director-password"
		blobstoreAgentPassword = "random-blobstore-agent-password"
		hmPassword = "random-hm-password"
		jobPropertiesManifestBuilder = manifests.NewJobPropertiesManifestBuilder(
			natsUsername,
			postgresUsername,
			registryUsername,
			blobstoreDirectorUsername,
			blobstoreAgentUsername,
			hmUsername,
			natsPassword,
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
				manifests.NATSJobProperties{
					Address:  "127.0.0.1",
					User:     natsUsername,
					Password: natsPassword,
				}))
		})
	})

	Describe("Postgres", func() {
		It("returns job properties for Postgres", func() {
			postgres := jobPropertiesManifestBuilder.Postgres()
			Expect(postgres).To(Equal(manifests.PostgresProperties{
				User:     postgresUsername,
				Password: postgresPassword,
			}))
		})
	})

	Describe("Registry", func() {
		It("returns job properties for Registry", func() {
			registry := jobPropertiesManifestBuilder.Registry()
			Expect(registry).To(Equal(manifests.RegistryJobProperties{
				Address:  "10.0.0.6",
				Host:     "10.0.0.6",
				Username: registryUsername,
				Password: registryPassword,
				DB: manifests.RegistryPostgresProperties{
					User:     postgresUsername,
					Password: postgresPassword,
					Database: "bosh",
				},
				HTTP: manifests.HTTPProperties{
					User:     registryUsername,
					Password: registryPassword,
				},
			}))
		})
	})

	Describe("Blobstore", func() {
		It("returns job properties for Blobstore", func() {
			blobstore := jobPropertiesManifestBuilder.Blobstore()
			Expect(blobstore).To(Equal(manifests.BlobstoreJobProperties{
				Address: "10.0.0.6",
				Director: manifests.Credentials{
					User:     blobstoreDirectorUsername,
					Password: blobstoreDirectorPassword,
				},
				Agent: manifests.Credentials{
					User:     blobstoreAgentUsername,
					Password: blobstoreAgentPassword,
				},
			}))
		})
	})

	Describe("Director", func() {
		It("returns job properties for Director", func() {
			director := jobPropertiesManifestBuilder.Director(manifests.ManifestProperties{
				DirectorName:     "my-bosh",
				DirectorUsername: "bosh-username",
				DirectorPassword: "bosh-password",
				SSLKeyPair: ssl.KeyPair{
					Certificate: []byte("some-ssl-cert"),
					PrivateKey:  []byte("some-ssl-key"),
				},
			})
			Expect(director).To(Equal(manifests.DirectorJobProperties{
				Address:          "127.0.0.1",
				Name:             "my-bosh",
				CPIJob:           "aws_cpi",
				MaxThreads:       10,
				EnablePostDeploy: true,
				DB: manifests.PostgresProperties{
					User:     postgresUsername,
					Password: postgresPassword,
				},
				UserManagement: manifests.UserManagementProperties{
					Local: manifests.LocalProperties{
						Users: []manifests.UserProperties{
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
				SSL: manifests.SSLProperties{
					Cert: "some-ssl-cert",
					Key:  "some-ssl-key",
				},
			}))
		})
	})

	Describe("HM", func() {
		It("returns job properties for HM", func() {
			hm := jobPropertiesManifestBuilder.HM()
			Expect(hm).To(Equal(manifests.HMJobProperties{
				DirectorAccount: manifests.Credentials{
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
			Expect(agent).To(Equal(manifests.AgentProperties{
				MBus: "nats://random-nats-username:random-nats-password@10.0.0.6:4222",
			}))
		})
	})
})
