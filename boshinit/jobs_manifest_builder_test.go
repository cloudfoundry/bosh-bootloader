package boshinit_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
)

var _ = Describe("JobsManifestBuilder", func() {
	var (
		jobsManifestBuilder boshinit.JobsManifestBuilder
		uuidGenerator       *fakes.UUIDGenerator
	)

	BeforeEach(func() {
		uuidGenerator = &fakes.UUIDGenerator{}
		jobsManifestBuilder = boshinit.NewJobsManifestBuilder(uuidGenerator)
	})

	Describe("Build", func() {
		BeforeEach(func() {
			uuidGenerator.GenerateCall.Returns = []fakes.GenerateReturn{
				{String: "randomly-generated-nats-password"},
				{String: "randomly-generated-redis-password"},
				{String: "randomly-generated-postgres-password"},
				{String: "randomly-generated-registry-password"},
				{String: "randomly-generated-blobstore-director-password"},
				{String: "randomly-generated-blobstore-agent-password"},
				{String: "randomly-generated-hm-password"},
			}
		})

		It("returns all jobs for manifest", func() {
			jobs, _, err := jobsManifestBuilder.Build(boshinit.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
			})

			Expect(err).NotTo(HaveOccurred())
			job := jobs[0]

			Expect(jobs).To(HaveLen(1))
			Expect(job.Name).To(Equal("bosh"))
			Expect(job.Instances).To(Equal(1))
			Expect(job.ResourcePool).To(Equal("vms"))
			Expect(job.PersistentDiskPool).To(Equal("disks"))

			Expect(job.Templates).To(ConsistOf([]boshinit.Template{
				{Name: "nats", Release: "bosh"},
				{Name: "redis", Release: "bosh"},
				{Name: "postgres", Release: "bosh"},
				{Name: "blobstore", Release: "bosh"},
				{Name: "director", Release: "bosh"},
				{Name: "health_monitor", Release: "bosh"},
				{Name: "registry", Release: "bosh"},
				{Name: "aws_cpi", Release: "bosh-aws-cpi"},
			}))

			Expect(job.Networks).To(ConsistOf([]boshinit.JobNetwork{
				{
					Name:      "private",
					StaticIPs: []string{"10.0.0.6"},
					Default:   []string{"dns", "gateway"},
				},
				{
					Name:      "public",
					StaticIPs: []string{"some-elastic-ip"},
				},
			}))

			Expect(job.Properties.NATS.User).To(Equal("nats"))
			Expect(job.Properties.Redis.Address).To(Equal("127.0.0.1"))
			Expect(job.Properties.Postgres.User).To(Equal("postgres"))
			Expect(job.Properties.Registry.Username).To(Equal("admin"))
			Expect(job.Properties.Blobstore.Provider).To(Equal("dav"))
			Expect(job.Properties.Director.Name).To(Equal("my-bosh"))
			Expect(job.Properties.HM.ResurrectorEnabled).To(Equal(true))
			Expect(job.Properties.AWS.AccessKeyId).To(Equal("some-access-key-id"))
			Expect(job.Properties.AWS.SecretAccessKey).To(Equal("some-secret-access-key"))
			Expect(job.Properties.AWS.Region).To(Equal("some-region"))
			Expect(job.Properties.AWS.DefaultKeyName).To(Equal("some-key-name"))
			Expect(job.Properties.Agent.MBus).To(Equal("nats://nats:randomly-generated-nats-password@10.0.0.6:4222"))
			Expect(job.Properties.NTP[0]).To(Equal("0.pool.ntp.org"))
		})

		It("returns manifest properties with new credentials", func() {
			_, manifestProperties, err := jobsManifestBuilder.Build(boshinit.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(manifestProperties.Credentials.NatsPassword).To(Equal("randomly-generated-nats-password"))
			Expect(manifestProperties.Credentials.RedisPassword).To(Equal("randomly-generated-redis-password"))
			Expect(manifestProperties.Credentials.PostgresPassword).To(Equal("randomly-generated-postgres-password"))
			Expect(manifestProperties.Credentials.RegistryPassword).To(Equal("randomly-generated-registry-password"))
			Expect(manifestProperties.Credentials.BlobstoreDirectorPassword).To(Equal("randomly-generated-blobstore-director-password"))
			Expect(manifestProperties.Credentials.BlobstoreAgentPassword).To(Equal("randomly-generated-blobstore-agent-password"))
			Expect(manifestProperties.Credentials.HMPassword).To(Equal("randomly-generated-hm-password"))
		})

		It("returns manifest and manifest properties with existing credentials", func() {
			jobs, manifestProperties, err := jobsManifestBuilder.Build(boshinit.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
				Credentials: boshinit.InternalCredentials{
					NatsPassword:              "some-persisted-nats-password",
					RedisPassword:             "some-persisted-redis-password",
					PostgresPassword:          "some-persisted-postgres-password",
					RegistryPassword:          "some-persisted-registry-password",
					BlobstoreDirectorPassword: "some-persisted-blobstore-director-password",
					BlobstoreAgentPassword:    "some-persisted-blobstore-agent-password",
					HMPassword:                "some-persisted-hm-password",
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(jobs).To(HaveLen(1))

			job := jobs[0]

			Expect(job.Properties.NATS.Password).To(Equal("some-persisted-nats-password"))
			Expect(job.Properties.Redis.Password).To(Equal("some-persisted-redis-password"))
			Expect(job.Properties.Postgres.Password).To(Equal("some-persisted-postgres-password"))
			Expect(job.Properties.Registry.Password).To(Equal("some-persisted-registry-password"))
			Expect(job.Properties.Registry.HTTP.Password).To(Equal("some-persisted-registry-password"))
			Expect(job.Properties.Blobstore.Director.Password).To(Equal("some-persisted-blobstore-director-password"))
			Expect(job.Properties.Blobstore.Agent.Password).To(Equal("some-persisted-blobstore-agent-password"))
			Expect(job.Properties.Director.UserManagement.Local.Users).To(ContainElement(boshinit.UserProperties{
				Name:     "hm",
				Password: "some-persisted-hm-password",
			}))
			Expect(job.Properties.HM.DirectorAccount.Password).To(Equal("some-persisted-hm-password"))

			Expect(manifestProperties.Credentials.NatsPassword).To(Equal("some-persisted-nats-password"))
			Expect(manifestProperties.Credentials.RedisPassword).To(Equal("some-persisted-redis-password"))
			Expect(manifestProperties.Credentials.PostgresPassword).To(Equal("some-persisted-postgres-password"))
			Expect(manifestProperties.Credentials.RegistryPassword).To(Equal("some-persisted-registry-password"))
			Expect(manifestProperties.Credentials.BlobstoreDirectorPassword).To(Equal("some-persisted-blobstore-director-password"))
			Expect(manifestProperties.Credentials.BlobstoreAgentPassword).To(Equal("some-persisted-blobstore-agent-password"))
			Expect(manifestProperties.Credentials.HMPassword).To(Equal("some-persisted-hm-password"))
		})

		It("uses the same password for NATS and the Agent", func() {
			jobs, _, err := jobsManifestBuilder.Build(boshinit.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
			})
			Expect(err).NotTo(HaveOccurred())

			job := jobs[0]

			Expect(job.Properties.Agent.MBus).To(Equal("nats://nats:randomly-generated-nats-password@10.0.0.6:4222"))
			Expect(job.Properties.NATS.Password).To(Equal("randomly-generated-nats-password"))
		})

		It("generates a password for redis", func() {
			jobs, _, err := jobsManifestBuilder.Build(boshinit.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
			})
			Expect(err).NotTo(HaveOccurred())

			job := jobs[0]

			Expect(job.Properties.Redis.Password).To(Equal("randomly-generated-redis-password"))
		})

		It("generates a password for postgres", func() {
			jobs, _, err := jobsManifestBuilder.Build(boshinit.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
			})
			Expect(err).NotTo(HaveOccurred())

			job := jobs[0]

			Expect(job.Properties.Postgres.Password).To(Equal("randomly-generated-postgres-password"))
			Expect(job.Properties.Registry.DB.Password).To(Equal("randomly-generated-postgres-password"))
			Expect(job.Properties.Director.DB.Password).To(Equal("randomly-generated-postgres-password"))
		})

		It("generates a password for blobstore director and agent", func() {
			jobs, _, err := jobsManifestBuilder.Build(boshinit.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
			})
			Expect(err).NotTo(HaveOccurred())

			job := jobs[0]

			Expect(job.Properties.Blobstore.Director.Password).To(Equal("randomly-generated-blobstore-director-password"))
			Expect(job.Properties.Blobstore.Agent.Password).To(Equal("randomly-generated-blobstore-agent-password"))
		})

		It("generates a password for health monitor", func() {
			jobs, _, err := jobsManifestBuilder.Build(boshinit.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
			})
			Expect(err).NotTo(HaveOccurred())

			job := jobs[0]

			Expect(job.Properties.HM.DirectorAccount.Password).To(Equal("randomly-generated-hm-password"))
			Expect(job.Properties.Director.UserManagement.Local.Users).To(ContainElement(
				boshinit.UserProperties{
					Name:     "hm",
					Password: "randomly-generated-hm-password",
				},
			))
		})

		Describe("failing to generate", func() {
			BeforeEach(func() {
				uuidGenerator.GenerateCall.Returns = []fakes.GenerateReturn{
					{String: "randomly-generated-nats-password"},
					{String: "randomly-generated-redis-password"},
					{String: "randomly-generated-postgres-password"},
					{String: "randomly-generated-registry-password"},
					{String: "randomly-generated-blobstore-director-password"},
					{String: "randomly-generated-blobstore-agent-password"},
					{String: "randomly-generated-hm-password"},
				}
			})

			Describe("the nats password", func() {
				BeforeEach(func() {
					uuidGenerator.GenerateCall.Returns[0] = fakes.GenerateReturn{
						Error: fmt.Errorf("error generating password"),
					}
				})

				It("forwards the error", func() {
					_, _, err := jobsManifestBuilder.Build(boshinit.ManifestProperties{})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error generating password"))
				})
			})

			Describe("the redis password", func() {
				BeforeEach(func() {
					uuidGenerator.GenerateCall.Returns[1] = fakes.GenerateReturn{
						Error: fmt.Errorf("error generating password"),
					}
				})

				It("forwards the error", func() {
					_, _, err := jobsManifestBuilder.Build(boshinit.ManifestProperties{})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error generating password"))

				})
			})

			Describe("the postgres password", func() {
				BeforeEach(func() {
					uuidGenerator.GenerateCall.Returns[2] = fakes.GenerateReturn{
						Error: fmt.Errorf("error generating password"),
					}
				})

				It("forwards the error", func() {
					_, _, err := jobsManifestBuilder.Build(boshinit.ManifestProperties{})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error generating password"))

				})
			})

			Describe("the registry password", func() {
				BeforeEach(func() {
					uuidGenerator.GenerateCall.Returns[3] = fakes.GenerateReturn{
						Error: fmt.Errorf("error generating password"),
					}
				})

				It("forwards the error", func() {
					_, _, err := jobsManifestBuilder.Build(boshinit.ManifestProperties{})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error generating password"))

				})
			})

			Describe("the blobstore director password", func() {
				BeforeEach(func() {
					uuidGenerator.GenerateCall.Returns[4] = fakes.GenerateReturn{
						Error: fmt.Errorf("error generating password"),
					}
				})

				It("forwards the error", func() {
					_, _, err := jobsManifestBuilder.Build(boshinit.ManifestProperties{})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error generating password"))

				})
			})

			Describe("the blobstore agent password", func() {
				BeforeEach(func() {
					uuidGenerator.GenerateCall.Returns[5] = fakes.GenerateReturn{
						Error: fmt.Errorf("error generating password"),
					}
				})

				It("forwards the error", func() {
					_, _, err := jobsManifestBuilder.Build(boshinit.ManifestProperties{})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error generating password"))

				})
			})

			Describe("the hm password", func() {
				BeforeEach(func() {
					uuidGenerator.GenerateCall.Returns[6] = fakes.GenerateReturn{
						Error: fmt.Errorf("error generating password"),
					}
				})

				It("forwards the error", func() {
					_, _, err := jobsManifestBuilder.Build(boshinit.ManifestProperties{})
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("error generating password"))

				})
			})
		})
	})
})
