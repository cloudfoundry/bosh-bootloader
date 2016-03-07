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

					Properties: map[string]interface{}{
						"nats": boshinit.NATSJobProperties{
							Address:  "127.0.0.1",
							User:     "nats",
							Password: "nats-password",
						},

						"redis": boshinit.RedisJobProperties{
							ListenAddress: "127.0.0.1",
							Address:       "127.0.0.1",
							Password:      "redis-password",
						},

						"postgres": boshinit.PostgresJobProperties{
							ListenAddress: "127.0.0.1",
							Host:          "127.0.0.1",
							User:          "postgres",
							Password:      "postgres-password",
							Database:      "bosh",
							Adapter:       "postgres",
						},

						"registry": boshinit.RegistryJobProperties{
							Address:  "10.0.0.6",
							Host:     "10.0.0.6",
							Username: "admin",
							Password: "admin",
							Port:     25777,
							DB: boshinit.DBProperties{
								ListenAddress: "127.0.0.1",
								Host:          "127.0.0.1",
								User:          "postgres",
								Password:      "postgres-password",
								Database:      "bosh",
								Adapter:       "postgres",
							},
							HTTP: boshinit.HTTPProperties{
								User:     "admin",
								Password: "admin",
								Port:     25777,
							},
						},

						"blobstore": boshinit.BlobstoreJobProperties{
							Address:  "10.0.0.6",
							Port:     25250,
							Provider: "dav",
							Director: boshinit.Credentials{
								User:     "director",
								Password: "director-password",
							},
							Agent: boshinit.Credentials{
								User:     "agent",
								Password: "agent-password",
							},
						},

						"director": boshinit.DirectorJobProperties{
							Address:    "127.0.0.1",
							Name:       "my-bosh",
							CPIJob:     "aws_cpi",
							MaxThreads: 10,
							DB: boshinit.DBProperties{
								ListenAddress: "127.0.0.1",
								Host:          "127.0.0.1",
								User:          "postgres",
								Password:      "postgres-password",
								Database:      "bosh",
								Adapter:       "postgres",
							},
							UserManagement: boshinit.UserManagementProperties{
								Provider: "local",
								Local: boshinit.LocalProperties{
									Users: []boshinit.UserProperties{
										{
											Name:     "admin",
											Password: "admin",
										},
										{
											Name:     "hm",
											Password: "hm-password",
										},
									},
								},
							},
						},

						"hm": boshinit.HMJobProperties{
							DirectorAccount: boshinit.Credentials{
								User:     "hm",
								Password: "hm-password",
							},
							ResurrectorEnabled: true,
						},

						"aws": boshinit.AWSJobProperties{
							AccessKeyId:           "ACCESS-KEY-ID",
							SecretAccessKey:       "SECRET-ACCESS-KEY",
							DefaultKeyName:        "bosh",
							DefaultSecurityGroups: []string{"bosh"},
							Region:                "REGION",
						},

						"agent": boshinit.AgentJobProperties{
							MBus: "nats://nats:nats-password@10.0.0.6:4222",
						},

						"ntp": []string{"0.pool.ntp.org", "1.pool.ntp.org"},
					},
				},
			}))
		})
	})
})
