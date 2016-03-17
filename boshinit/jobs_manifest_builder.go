package boshinit

type JobsManifestBuilder struct {
	stringGenerator stringGenerator
}

func NewJobsManifestBuilder(stringGenerator stringGenerator) JobsManifestBuilder {
	return JobsManifestBuilder{
		stringGenerator: stringGenerator,
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
	var passwords = map[string]*string{}
	passwords["nats-"] = &manifestProperties.Credentials.NatsPassword
	passwords["redis-"] = &manifestProperties.Credentials.RedisPassword
	passwords["postgres-"] = &manifestProperties.Credentials.PostgresPassword
	passwords["registry-"] = &manifestProperties.Credentials.RegistryPassword
	passwords["blobstore-director-"] = &manifestProperties.Credentials.BlobstoreDirectorPassword
	passwords["blobstore-agent-"] = &manifestProperties.Credentials.BlobstoreAgentPassword
	passwords["hm-"] = &manifestProperties.Credentials.HMPassword

	for key, value := range passwords {
		if *value == "" {
			generatedString, err := j.stringGenerator.Generate(key, PASSWORD_LENGTH)
			if err != nil {
				return manifestProperties, err
			}
			*value = generatedString
		}
	}

	return manifestProperties, nil
}
