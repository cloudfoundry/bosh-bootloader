package boshinit

type ManifestBuilder struct {
	logger logger
}

type ManifestProperties struct {
	SubnetID         string
	AvailabilityZone string
	ElasticIP        string
	AccessKeyID      string
	SecretAccessKey  string
	DefaultKeyName   string
	Region           string
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
		ResourcePools: resourcePoolsManifestBuilder.Build(manifestProperties),
		DiskPools:     diskPoolsManifestBuilder.Build(),
		Networks:      networksManifestBuilder.Build(manifestProperties),
		Jobs:          jobsManifestBuilder.Build(manifestProperties),
		CloudProvider: cloudProviderManifestBuilder.Build(manifestProperties),
	}
}
