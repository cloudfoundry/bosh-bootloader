package boshinit

type JobsManifestBuilder struct {
	uuidGenerator UUIDGenerator
}

func NewJobsManifestBuilder(uuidGenerator UUIDGenerator) JobsManifestBuilder {
	return JobsManifestBuilder{
		uuidGenerator: uuidGenerator,
	}
}

func (j JobsManifestBuilder) Build(manifestProperties ManifestProperties) ([]Job, error) {
	sharedPropertiesManifestBuilder := NewSharedPropertiesManifestBuilder()

	natsPassword, err := j.uuidGenerator.Generate()
	if err != nil {
		return nil, err
	}

	redisPassword, err := j.uuidGenerator.Generate()
	if err != nil {
		return nil, err
	}

	postgresPassword, err := j.uuidGenerator.Generate()
	if err != nil {
		return nil, err
	}

	registryPassword, err := j.uuidGenerator.Generate()
	if err != nil {
		return nil, err
	}

	blobstoreDirectorPassword, err := j.uuidGenerator.Generate()
	if err != nil {
		return nil, err
	}

	blobstoreAgentPassword, err := j.uuidGenerator.Generate()
	if err != nil {
		return nil, err
	}

	hmPassword, err := j.uuidGenerator.Generate()
	if err != nil {
		return nil, err
	}

	jobPropertiesManifestBuilder := NewJobPropertiesManifestBuilder(
		natsPassword,
		redisPassword,
		postgresPassword,
		registryPassword,
		blobstoreDirectorPassword,
		blobstoreAgentPassword,
		hmPassword,
	)

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
				Postgres:  jobPropertiesManifestBuilder.Postgres(),
				Registry:  jobPropertiesManifestBuilder.Registry(),
				Blobstore: jobPropertiesManifestBuilder.Blobstore(),
				Director:  jobPropertiesManifestBuilder.Director(manifestProperties),
				HM:        jobPropertiesManifestBuilder.HM(),
				AWS:       sharedPropertiesManifestBuilder.AWS(manifestProperties),
				Agent:     jobPropertiesManifestBuilder.Agent(),
				NTP:       sharedPropertiesManifestBuilder.NTP(),
			},
		},
	}, nil
}
