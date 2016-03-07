package boshinit

type ManifestBuilder struct{}

func NewManifestBuilder() ManifestBuilder {
	return ManifestBuilder{}
}

func (m ManifestBuilder) Build() Manifest {
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
		Networks:      networksManifestBuilder.Build(),
		Jobs:          jobsManifestBuilder.Build(),
		CloudProvider: cloudProviderManifestBuilder.Build(),
	}
}
