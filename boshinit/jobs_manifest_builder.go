package boshinit

type JobsManifestBuilder struct{}

func NewJobsManifestBuilder() JobsManifestBuilder {
	return JobsManifestBuilder{}
}

func (r JobsManifestBuilder) Build() []Job {
	jobPropertiesManifestBuilder := NewJobPropertiesManifestBuilder()

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

			Properties: JobProperties{
				NATS:      jobPropertiesManifestBuilder.NATS(),
				Redis:     jobPropertiesManifestBuilder.Redis(),
				Postgres:  jobPropertiesManifestBuilder.Postgres(),
				Registry:  jobPropertiesManifestBuilder.Registry(),
				Blobstore: jobPropertiesManifestBuilder.Blobstore(),
				Director:  jobPropertiesManifestBuilder.Director(),
				HM:        jobPropertiesManifestBuilder.HM(),
				AWS:       jobPropertiesManifestBuilder.AWS(),
				Agent:     jobPropertiesManifestBuilder.Agent(),
				NTP:       jobPropertiesManifestBuilder.NTP(),
			},
		},
	}
}
