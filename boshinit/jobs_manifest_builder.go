package boshinit

type JobsManifestBuilder struct {
	uuidGenerator UUIDGenerator
}

func NewJobsManifestBuilder(uuidGenerator UUIDGenerator) JobsManifestBuilder {
	return JobsManifestBuilder{
		uuidGenerator: uuidGenerator,
	}
}

func (j JobsManifestBuilder) Build(manifestProperties ManifestProperties) ([]Job, ManifestProperties, error) {
	sharedPropertiesManifestBuilder := NewSharedPropertiesManifestBuilder()

	manifestProperties, err := j.generateInternalPasswords(manifestProperties)
	if err != nil {
		return nil, ManifestProperties{}, err
	}

	jobPropertiesManifestBuilder := NewJobPropertiesManifestBuilder(
		manifestProperties.Credentials.NatsPassword,
		manifestProperties.Credentials.RedisPassword,
		manifestProperties.Credentials.PostgresPassword,
		manifestProperties.Credentials.RegistryPassword,
		manifestProperties.Credentials.BlobstoreDirectorPassword,
		manifestProperties.Credentials.BlobstoreAgentPassword,
		manifestProperties.Credentials.HMPassword,
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
	}, manifestProperties, nil
}

func (j JobsManifestBuilder) generateInternalPasswords(manifestProperties ManifestProperties) (ManifestProperties, error) {
	var err error

	if manifestProperties.Credentials.NatsPassword == "" {
		manifestProperties.Credentials.NatsPassword, err = j.uuidGenerator.Generate()
		if err != nil {
			return ManifestProperties{}, err
		}
	}

	if manifestProperties.Credentials.RedisPassword == "" {
		manifestProperties.Credentials.RedisPassword, err = j.uuidGenerator.Generate()
		if err != nil {
			return ManifestProperties{}, err
		}
	}

	if manifestProperties.Credentials.PostgresPassword == "" {
		manifestProperties.Credentials.PostgresPassword, err = j.uuidGenerator.Generate()
		if err != nil {
			return ManifestProperties{}, err
		}
	}

	if manifestProperties.Credentials.RegistryPassword == "" {
		manifestProperties.Credentials.RegistryPassword, err = j.uuidGenerator.Generate()
		if err != nil {
			return ManifestProperties{}, err
		}
	}

	if manifestProperties.Credentials.BlobstoreDirectorPassword == "" {
		manifestProperties.Credentials.BlobstoreDirectorPassword, err = j.uuidGenerator.Generate()
		if err != nil {
			return ManifestProperties{}, err
		}
	}

	if manifestProperties.Credentials.BlobstoreAgentPassword == "" {
		manifestProperties.Credentials.BlobstoreAgentPassword, err = j.uuidGenerator.Generate()
		if err != nil {
			return ManifestProperties{}, err
		}
	}

	if manifestProperties.Credentials.HMPassword == "" {
		manifestProperties.Credentials.HMPassword, err = j.uuidGenerator.Generate()
		if err != nil {
			return ManifestProperties{}, err
		}
	}

	return manifestProperties, nil
}
