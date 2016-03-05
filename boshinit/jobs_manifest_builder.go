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

			Properties: map[string]JobProperties{
				"nats": JobProperties{
					Address:  "127.0.0.1",
					User:     "nats",
					Password: "nats-password",
				},

				"redis": JobProperties{
					ListenAddress: "127.0.0.1",
					Address:       "127.0.0.1",
					Password:      "redis-password",
				},

				"postgres": JobProperties{
					ListenAddress: "127.0.0.1",
					Host:          "127.0.0.1",
					User:          "postgres",
					Password:      "postgres-password",
					Database:      "bosh",
					Adapter:       "postgres",
				},

				"registry": JobProperties{
					Address: "10.0.0.6",
				},
			},
		},
	}
}
