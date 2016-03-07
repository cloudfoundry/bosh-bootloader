package boshinit

type JobsManifestBuilder struct{}

func NewJobsManifestBuilder() JobsManifestBuilder {
	return JobsManifestBuilder{}
}

func (r JobsManifestBuilder) Build() []Job {
	return []Job{
		{
			Name:               "bosh",
			Instances:          1,
			ResourcePool:       "vms",
			PersistentDiskPool: "disks",

			Templates: []Template{
				{Name: "nats", Release: "bosh"},
				{Name: "redis", Release: "bosh"},
				{Name: "postgres", Release: "bosh"},
				{Name: "blobstore", Release: "bosh"},
				{Name: "director", Release: "bosh"},
				{Name: "health_monitor", Release: "bosh"},
				{Name: "registry", Release: "bosh"},
				{Name: "aws_cpi", Release: "bosh-aws-cpi"},
			},

			Networks: []JobNetwork{
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
				"nats": NATSJobProperties{
					Address:  "127.0.0.1",
					User:     "nats",
					Password: "nats-password",
				},

				"redis": RedisJobProperties{
					ListenAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Password:      "redis-password",
				},

				"postgres": PostgresJobProperties{
					ListenAddress: "127.0.0.1",
					Host:          "127.0.0.1",
					User:          "postgres",
					Password:      "postgres-password",
					Database:      "bosh",
					Adapter:       "postgres",
				},

				"registry": RegistryJobProperties{
					Address:  "10.0.0.6",
					Host:     "10.0.0.6",
					Username: "admin",
					Password: "admin",
					Port:     25777,
					DB: DBProperties{
						ListenAddress: "127.0.0.1",
						Host:          "127.0.0.1",
						User:          "postgres",
						Password:      "postgres-password",
						Database:      "bosh",
						Adapter:       "postgres",
					},
					HTTP: HTTPProperties{
						User:     "admin",
						Password: "admin",
						Port:     25777,
					},
				},

				"blobstore": BlobstoreJobProperties{
					Address:  "10.0.0.6",
					Port:     25250,
					Provider: "dav",
					Director: Credentials{
						User:     "director",
						Password: "director-password",
					},
					Agent: Credentials{
						User:     "agent",
						Password: "agent-password",
					},
				},

				"director": DirectorJobProperties{
					Address:    "127.0.0.1",
					Name:       "my-bosh",
					CPIJob:     "aws_cpi",
					MaxThreads: 10,
					DB: DBProperties{
						ListenAddress: "127.0.0.1",
						Host:          "127.0.0.1",
						User:          "postgres",
						Password:      "postgres-password",
						Database:      "bosh",
						Adapter:       "postgres",
					},
					UserManagement: UserManagementProperties{
						Provider: "local",
						Local: LocalProperties{
							Users: []UserProperties{
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

				"hm": HMJobProperties{
					DirectorAccount: Credentials{
						User:     "hm",
						Password: "hm-password",
					},
					ResurrectorEnabled: true,
				},

				"aws": AWSJobProperties{
					AccessKeyId:           "ACCESS-KEY-ID",
					SecretAccessKey:       "SECRET-ACCESS-KEY",
					DefaultKeyName:        "bosh",
					DefaultSecurityGroups: []string{"bosh"},
					Region:                "REGION",
				},

				"agent": AgentJobProperties{
					MBus: "nats://nats:nats-password@10.0.0.6:4222",
				},

				"ntp": []string{"0.pool.ntp.org", "1.pool.ntp.org"},
			},
		},
	}
}
