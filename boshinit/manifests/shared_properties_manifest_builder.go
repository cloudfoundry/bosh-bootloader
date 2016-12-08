package manifests

type SharedPropertiesManifestBuilder struct {
}

func NewSharedPropertiesManifestBuilder() *SharedPropertiesManifestBuilder {
	return &SharedPropertiesManifestBuilder{}
}

func (SharedPropertiesManifestBuilder) AWS(manifestProperties ManifestProperties) AWSProperties {
	return AWSProperties{
		AccessKeyId:           manifestProperties.AWS.AccessKeyID,
		SecretAccessKey:       manifestProperties.AWS.SecretAccessKey,
		DefaultKeyName:        manifestProperties.AWS.DefaultKeyName,
		DefaultSecurityGroups: []string{manifestProperties.AWS.SecurityGroup},
		Region:                manifestProperties.AWS.Region,
	}
}

func (SharedPropertiesManifestBuilder) Google(manifestProperties ManifestProperties) GoogleProperties {
	return GoogleProperties{
		Project: manifestProperties.GCP.Project,
		JsonKey: manifestProperties.GCP.JsonKey,
	}
}
