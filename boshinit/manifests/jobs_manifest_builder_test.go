package manifests_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit/manifests"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"
)

var _ = Describe("JobsManifestBuilder", func() {
	var (
		jobsManifestBuilder manifests.JobsManifestBuilder
		stringGenerator     *fakes.StringGenerator
	)

	BeforeEach(func() {
		stringGenerator = &fakes.StringGenerator{}
		jobsManifestBuilder = manifests.NewJobsManifestBuilder(stringGenerator)
	})

	Describe("Build", func() {
		BeforeEach(func() {
			stringGenerator.GenerateCall.Stub = func(prefix string, length int) (string, error) {
				if length != 15 {
					return "", errors.New("wrong length passed to string generator")
				}
				return fmt.Sprintf("%s%s", prefix, "some-random-string"), nil
			}
		})

		It("returns all jobs for manifest", func() {
			jobs, _, err := jobsManifestBuilder.Build(manifests.ManifestProperties{
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

			Expect(job.Templates).To(ConsistOf([]manifests.Template{
				{Name: "nats", Release: "bosh"},
				{Name: "postgres", Release: "bosh"},
				{Name: "blobstore", Release: "bosh"},
				{Name: "director", Release: "bosh"},
				{Name: "health_monitor", Release: "bosh"},
				{Name: "registry", Release: "bosh"},
				{Name: "aws_cpi", Release: "bosh-aws-cpi"},
				{Name: "powerdns", Release: "bosh"},
			}))

			Expect(job.Networks).To(ConsistOf([]manifests.JobNetwork{
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

			Expect(job.Properties.NATS.User).To(Equal("nats-user-some-random-string"))
			Expect(job.Properties.Postgres.User).To(Equal("postgres-user-some-random-string"))
			Expect(job.Properties.Registry.Username).To(Equal("registry-user-some-random-string"))
			Expect(job.Properties.Blobstore.Provider).To(Equal("dav"))
			Expect(job.Properties.Director.Name).To(Equal("my-bosh"))
			Expect(job.Properties.HM.ResurrectorEnabled).To(Equal(true))
			Expect(job.Properties.AWS.AccessKeyId).To(Equal("some-access-key-id"))
			Expect(job.Properties.AWS.SecretAccessKey).To(Equal("some-secret-access-key"))
			Expect(job.Properties.AWS.Region).To(Equal("some-region"))
			Expect(job.Properties.AWS.DefaultKeyName).To(Equal("some-key-name"))
			Expect(job.Properties.Agent.MBus).To(Equal("nats://nats-user-some-random-string:nats-some-random-string@10.0.0.6:4222"))
			Expect(job.Properties.NTP[0]).To(Equal("0.pool.ntp.org"))
		})

		It("returns manifest properties with new credentials", func() {
			_, manifestProperties, err := jobsManifestBuilder.Build(manifests.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(manifestProperties.Credentials.NatsUsername).To(Equal("nats-user-some-random-string"))
			Expect(manifestProperties.Credentials.PostgresUsername).To(Equal("postgres-user-some-random-string"))
			Expect(manifestProperties.Credentials.RegistryUsername).To(Equal("registry-user-some-random-string"))
			Expect(manifestProperties.Credentials.BlobstoreDirectorUsername).To(Equal("blobstore-director-user-some-random-string"))
			Expect(manifestProperties.Credentials.BlobstoreAgentUsername).To(Equal("blobstore-agent-user-some-random-string"))
			Expect(manifestProperties.Credentials.HMUsername).To(Equal("hm-user-some-random-string"))
			Expect(manifestProperties.Credentials.NatsPassword).To(Equal("nats-some-random-string"))
			Expect(manifestProperties.Credentials.PostgresPassword).To(Equal("postgres-some-random-string"))
			Expect(manifestProperties.Credentials.RegistryPassword).To(Equal("registry-some-random-string"))
			Expect(manifestProperties.Credentials.BlobstoreDirectorPassword).To(Equal("blobstore-director-some-random-string"))
			Expect(manifestProperties.Credentials.BlobstoreAgentPassword).To(Equal("blobstore-agent-some-random-string"))
			Expect(manifestProperties.Credentials.HMPassword).To(Equal("hm-some-random-string"))
		})

		It("returns manifest and manifest properties with existing credentials", func() {
			jobs, manifestProperties, err := jobsManifestBuilder.Build(manifests.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
				Credentials: manifests.InternalCredentials{
					NatsUsername:              "some-persisted-nats-username",
					PostgresUsername:          "some-persisted-postgres-username",
					RegistryUsername:          "some-persisted-registry-username",
					BlobstoreDirectorUsername: "some-persisted-blobstore-director-username",
					BlobstoreAgentUsername:    "some-persisted-blobstore-agent-username",
					HMUsername:                "some-persisted-hm-username",
					NatsPassword:              "some-persisted-nats-password",
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

			Expect(job.Properties.NATS.User).To(Equal("some-persisted-nats-username"))
			Expect(job.Properties.Postgres.User).To(Equal("some-persisted-postgres-username"))
			Expect(job.Properties.Registry.Username).To(Equal("some-persisted-registry-username"))
			Expect(job.Properties.Registry.HTTP.User).To(Equal("some-persisted-registry-username"))
			Expect(job.Properties.Blobstore.Director.User).To(Equal("some-persisted-blobstore-director-username"))
			Expect(job.Properties.Blobstore.Agent.User).To(Equal("some-persisted-blobstore-agent-username"))
			Expect(job.Properties.Director.UserManagement.Local.Users).To(ContainElement(manifests.UserProperties{
				Name:     "some-persisted-hm-username",
				Password: "some-persisted-hm-password",
			}))
			Expect(job.Properties.HM.DirectorAccount.User).To(Equal("some-persisted-hm-username"))

			Expect(manifestProperties.Credentials.NatsUsername).To(Equal("some-persisted-nats-username"))
			Expect(manifestProperties.Credentials.PostgresUsername).To(Equal("some-persisted-postgres-username"))
			Expect(manifestProperties.Credentials.RegistryUsername).To(Equal("some-persisted-registry-username"))
			Expect(manifestProperties.Credentials.BlobstoreDirectorUsername).To(Equal("some-persisted-blobstore-director-username"))
			Expect(manifestProperties.Credentials.BlobstoreAgentUsername).To(Equal("some-persisted-blobstore-agent-username"))
			Expect(manifestProperties.Credentials.HMUsername).To(Equal("some-persisted-hm-username"))

			Expect(job.Properties.NATS.Password).To(Equal("some-persisted-nats-password"))
			Expect(job.Properties.Postgres.Password).To(Equal("some-persisted-postgres-password"))
			Expect(job.Properties.Registry.Password).To(Equal("some-persisted-registry-password"))
			Expect(job.Properties.Registry.HTTP.Password).To(Equal("some-persisted-registry-password"))
			Expect(job.Properties.Blobstore.Director.Password).To(Equal("some-persisted-blobstore-director-password"))
			Expect(job.Properties.Blobstore.Agent.Password).To(Equal("some-persisted-blobstore-agent-password"))
			Expect(job.Properties.HM.DirectorAccount.Password).To(Equal("some-persisted-hm-password"))

			Expect(manifestProperties.Credentials.NatsPassword).To(Equal("some-persisted-nats-password"))
			Expect(manifestProperties.Credentials.PostgresPassword).To(Equal("some-persisted-postgres-password"))
			Expect(manifestProperties.Credentials.RegistryPassword).To(Equal("some-persisted-registry-password"))
			Expect(manifestProperties.Credentials.BlobstoreDirectorPassword).To(Equal("some-persisted-blobstore-director-password"))
			Expect(manifestProperties.Credentials.BlobstoreAgentPassword).To(Equal("some-persisted-blobstore-agent-password"))
			Expect(manifestProperties.Credentials.HMPassword).To(Equal("some-persisted-hm-password"))
		})

		It("uses the same credentials for NATS and the Agent", func() {
			jobs, _, err := jobsManifestBuilder.Build(manifests.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
			})
			Expect(err).NotTo(HaveOccurred())

			job := jobs[0]

			Expect(job.Properties.Agent.MBus).To(Equal("nats://nats-user-some-random-string:nats-some-random-string@10.0.0.6:4222"))
			Expect(job.Properties.NATS.User).To(Equal("nats-user-some-random-string"))
			Expect(job.Properties.NATS.Password).To(Equal("nats-some-random-string"))
		})

		It("generates a password for postgres", func() {
			jobs, _, err := jobsManifestBuilder.Build(manifests.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
			})
			Expect(err).NotTo(HaveOccurred())

			job := jobs[0]

			Expect(job.Properties.Postgres.User).To(Equal("postgres-user-some-random-string"))
			Expect(job.Properties.Registry.DB.User).To(Equal("postgres-user-some-random-string"))
			Expect(job.Properties.Director.DB.User).To(Equal("postgres-user-some-random-string"))
			Expect(job.Properties.DNS.DB.User).To(Equal("postgres-user-some-random-string"))
			Expect(job.Properties.Postgres.Password).To(Equal("postgres-some-random-string"))
			Expect(job.Properties.Registry.DB.Password).To(Equal("postgres-some-random-string"))
			Expect(job.Properties.Director.DB.Password).To(Equal("postgres-some-random-string"))
			Expect(job.Properties.DNS.DB.Password).To(Equal("postgres-some-random-string"))
		})

		It("generates a password for blobstore director and agent", func() {
			jobs, _, err := jobsManifestBuilder.Build(manifests.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
			})
			Expect(err).NotTo(HaveOccurred())

			job := jobs[0]

			Expect(job.Properties.Blobstore.Director.User).To(Equal("blobstore-director-user-some-random-string"))
			Expect(job.Properties.Blobstore.Agent.User).To(Equal("blobstore-agent-user-some-random-string"))
			Expect(job.Properties.Blobstore.Director.Password).To(Equal("blobstore-director-some-random-string"))
			Expect(job.Properties.Blobstore.Agent.Password).To(Equal("blobstore-agent-some-random-string"))
		})

		It("generates a password for health monitor", func() {
			jobs, _, err := jobsManifestBuilder.Build(manifests.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
			})
			Expect(err).NotTo(HaveOccurred())

			job := jobs[0]

			Expect(job.Properties.HM.DirectorAccount.User).To(Equal("hm-user-some-random-string"))
			Expect(job.Properties.HM.DirectorAccount.Password).To(Equal("hm-some-random-string"))
			Expect(job.Properties.Director.UserManagement.Local.Users).To(ContainElement(
				manifests.UserProperties{
					Name:     "hm-user-some-random-string",
					Password: "hm-some-random-string",
				},
			))
		})

		Context("failure cases", func() {
			It("returns an error when string generation fails", func() {
				stringGenerator.GenerateCall.Stub = nil
				stringGenerator.GenerateCall.Returns.Error = errors.New("string generation failed")

				_, _, err := jobsManifestBuilder.Build(manifests.ManifestProperties{})
				Expect(err).To(MatchError("string generation failed"))
			})
		})
	})
})
