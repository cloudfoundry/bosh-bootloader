package manifests

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

	manifestProperties, err := j.generateInternalCredentials(manifestProperties)
	if err != nil {
		return nil, ManifestProperties{}, err
	}

	jobPropertiesManifestBuilder := NewJobPropertiesManifestBuilder(
		manifestProperties.Credentials.NatsUsername,
		manifestProperties.Credentials.PostgresUsername,
		manifestProperties.Credentials.RegistryUsername,
		manifestProperties.Credentials.BlobstoreDirectorUsername,
		manifestProperties.Credentials.BlobstoreAgentUsername,
		manifestProperties.Credentials.HMUsername,
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

func (j JobsManifestBuilder) generateInternalCredentials(manifestProperties ManifestProperties) (ManifestProperties, error) {
	var credentials = map[string]*string{}
	credentials["nats-user-"] = &manifestProperties.Credentials.NatsUsername
	credentials["postgres-user-"] = &manifestProperties.Credentials.PostgresUsername
	credentials["registry-user-"] = &manifestProperties.Credentials.RegistryUsername
	credentials["blobstore-director-user-"] = &manifestProperties.Credentials.BlobstoreDirectorUsername
	credentials["blobstore-agent-user-"] = &manifestProperties.Credentials.BlobstoreAgentUsername
	credentials["hm-user-"] = &manifestProperties.Credentials.HMUsername
	credentials["nats-"] = &manifestProperties.Credentials.NatsPassword
	credentials["redis-"] = &manifestProperties.Credentials.RedisPassword
	credentials["postgres-"] = &manifestProperties.Credentials.PostgresPassword
	credentials["registry-"] = &manifestProperties.Credentials.RegistryPassword
	credentials["blobstore-director-"] = &manifestProperties.Credentials.BlobstoreDirectorPassword
	credentials["blobstore-agent-"] = &manifestProperties.Credentials.BlobstoreAgentPassword
	credentials["hm-"] = &manifestProperties.Credentials.HMPassword

	for key, value := range credentials {
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
