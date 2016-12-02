package manifests

type JobsManifestBuilder struct {
	stringGenerator stringGenerator
}

func NewJobsManifestBuilder(stringGenerator stringGenerator) JobsManifestBuilder {
	return JobsManifestBuilder{
		stringGenerator: stringGenerator,
	}
}

func (j JobsManifestBuilder) Build(iaas string, manifestProperties ManifestProperties) ([]Job, ManifestProperties, error) {
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
		manifestProperties.Credentials.PostgresPassword,
		manifestProperties.Credentials.RegistryPassword,
		manifestProperties.Credentials.BlobstoreDirectorPassword,
		manifestProperties.Credentials.BlobstoreAgentPassword,
		manifestProperties.Credentials.HMPassword,
	)

	cpiName, cpiRelease := getCPIJob(iaas)
	jobProperties := JobProperties{
		NATS:      jobPropertiesManifestBuilder.NATS(),
		Postgres:  jobPropertiesManifestBuilder.Postgres(),
		Registry:  jobPropertiesManifestBuilder.Registry(),
		Blobstore: jobPropertiesManifestBuilder.Blobstore(),
		Director:  jobPropertiesManifestBuilder.Director(iaas, manifestProperties),
		HM:        jobPropertiesManifestBuilder.HM(),
		Agent:     jobPropertiesManifestBuilder.Agent(),
	}

	if iaas == "aws" {
		jobProperties.AWS = sharedPropertiesManifestBuilder.AWS(manifestProperties)
	} else {
		jobProperties.Google = sharedPropertiesManifestBuilder.Google(manifestProperties)
	}

	return []Job{
		{
			Name:               "bosh",
			Instances:          1,
			ResourcePool:       "vms",
			PersistentDiskPool: "disks",

			Templates: []Template{
				{Name: "nats", Release: "bosh"},
				{Name: "postgres", Release: "bosh"},
				{Name: "blobstore", Release: "bosh"},
				{Name: "director", Release: "bosh"},
				{Name: "health_monitor", Release: "bosh"},
				{Name: "registry", Release: "bosh"},
				{Name: cpiName, Release: cpiRelease},
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

			Properties: jobProperties,
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

func getCPIJob(iaas string) (name, release string) {
	switch iaas {
	case "aws":
		return "aws_cpi", "bosh-aws-cpi"
	case "gcp":
		return "google_cpi", "bosh-google-cpi"
	default:
		return "", ""
	}
}
