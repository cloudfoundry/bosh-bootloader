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
			jobs := jobsManifestBuilder.Build()

			Expect(jobs).To(HaveLen(1))
			Expect(jobs).To(ConsistOf([]boshinit.Job{
				{
					Name:               "bosh",
					Instances:          1,
					ResourcePool:       "vms",
					PersistentDiskPool: "disks",

					Templates: []boshinit.Template{
						{Name: "nats", Release: "bosh"},
						{Name: "redis", Release: "bosh"},
						{Name: "postgres", Release: "bosh"},
						{Name: "blobstore", Release: "bosh"},
						{Name: "director", Release: "bosh"},
						{Name: "health_monitor", Release: "bosh"},
						{Name: "registry", Release: "bosh"},
						{Name: "aws_cpi", Release: "bosh-aws-cpi"},
					},

					Networks: []boshinit.JobNetwork{
						{
							Name:      "private",
							StaticIPs: []string{"10.0.0.6"},
							Default:   []string{"dns", "gateway"},
						},
						{
							Name:      "public",
							StaticIPs: []string{"ELASTIC-IP"},
						},
					},

					Properties: map[string]boshinit.JobProperties{
						"nats": boshinit.JobProperties{
							Address:  "127.0.0.1",
							User:     "nats",
							Password: "nats-password",
						},

						"redis": boshinit.JobProperties{
							ListenAddress: "127.0.0.1",
							Address:       "127.0.0.1",
							Password:      "redis-password",
						},

						"postgres": boshinit.JobProperties{
							ListenAddress: "127.0.0.1",
							Host:          "127.0.0.1",
							User:          "postgres",
							Password:      "postgres-password",
							Database:      "bosh",
							Adapter:       "postgres",
						},

						"registry": boshinit.JobProperties{
							Address: "10.0.0.6",
							Host:    "",
						},
					},
				},
			}))
		})
	})
})
