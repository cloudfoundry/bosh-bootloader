package boshinit_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
)

var _ = Describe("JobsManifestBuilder", func() {
	var jobsManifestBuilder boshinit.JobsManifestBuilder

	BeforeEach(func() {
		jobsManifestBuilder = boshinit.NewJobsManifestBuilder()
	})

	Describe("Build", func() {
		It("returns all jobs for manifest", func() {
			jobs := jobsManifestBuilder.Build(boshinit.ManifestProperties{
				ElasticIP:       "some-elastic-ip",
				AccessKeyID:     "some-access-key-id",
				SecretAccessKey: "some-secret-access-key",
				DefaultKeyName:  "some-key-name",
				Region:          "some-region",
			})

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
			Expect(job.Properties.Agent.MBus).To(Equal("nats://nats:nats-password@10.0.0.6:4222"))
			Expect(job.Properties.NTP[0]).To(Equal("0.pool.ntp.org"))
		})
	})
})
