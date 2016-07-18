package manifests

type SharedPropertiesManifestBuilder struct {
}

func NewSharedPropertiesManifestBuilder() *SharedPropertiesManifestBuilder {
	return &SharedPropertiesManifestBuilder{}
}

func (SharedPropertiesManifestBuilder) AWS(manifestProperties ManifestProperties) AWSProperties {
	return AWSProperties{
		AccessKeyId:           manifestProperties.AccessKeyID,
		SecretAccessKey:       manifestProperties.SecretAccessKey,
		DefaultKeyName:        manifestProperties.DefaultKeyName,
		DefaultSecurityGroups: []string{manifestProperties.SecurityGroup},
		Region:                manifestProperties.Region,
	}
}
