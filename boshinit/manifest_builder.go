package boshinit

type ManifestBuilder struct {
	logger logger
}

type ManifestProperties struct {
	SubnetID string
}

type logger interface {
	Step(message string)
}

func NewManifestBuilder(logger logger) ManifestBuilder {
	return ManifestBuilder{
		logger: logger,
	}
}

func (m ManifestBuilder) Build(manifestProperties ManifestProperties) Manifest {
	m.logger.Step("generating bosh-init manifest")

	releaseManifestBuilder := NewReleaseManifestBuilder()
	resourcePoolsManifestBuilder := NewResourcePoolsManifestBuilder()
	diskPoolsManifestBuilder := NewDiskPoolsManifestBuilder()
	networksManifestBuilder := NewNetworksManifestBuilder()
	jobsManifestBuilder := NewJobsManifestBuilder()
	cloudProviderManifestBuilder := NewCloudProviderManifestBuilder()

	return Manifest{
		Name:          "bosh",
		Releases:      releaseManifestBuilder.Build(),
		ResourcePools: resourcePoolsManifestBuilder.Build(),
		DiskPools:     diskPoolsManifestBuilder.Build(),
		Networks:      networksManifestBuilder.Build(manifestProperties.SubnetID),
		Jobs:          jobsManifestBuilder.Build(),
		CloudProvider: cloudProviderManifestBuilder.Build(),
	}
}
