package boshinit

type JobsManifestBuilder struct{}

func NewJobsManifestBuilder() JobsManifestBuilder {
	return JobsManifestBuilder{}
}

func (r JobsManifestBuilder) Build(manifestProperties ManifestProperties) []Job {
	jobPropertiesManifestBuilder := NewJobPropertiesManifestBuilder()
	sharedPropertiesManifestBuilder := NewSharedPropertiesManifestBuilder()

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
					StaticIPs: []string{manifestProperties.ElasticIP},
				},
			},

			Properties: JobProperties{
				NATS:      jobPropertiesManifestBuilder.NATS(),
				Redis:     jobPropertiesManifestBuilder.Redis(),
				Postgres:  sharedPropertiesManifestBuilder.Postgres(),
				Registry:  jobPropertiesManifestBuilder.Registry(),
				Blobstore: jobPropertiesManifestBuilder.Blobstore(),
				Director:  jobPropertiesManifestBuilder.Director(),
				HM:        jobPropertiesManifestBuilder.HM(),
				AWS:       sharedPropertiesManifestBuilder.AWS(manifestProperties),
				Agent:     jobPropertiesManifestBuilder.Agent(),
				NTP:       sharedPropertiesManifestBuilder.NTP(),
			},
		},
	}
}
