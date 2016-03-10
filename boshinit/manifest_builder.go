package boshinit

import "github.com/pivotal-cf-experimental/bosh-bootloader/ssl"

type ManifestBuilder struct {
	logger              logger
	sslKeyPairGenerator sslKeyPairGenerator
}

type ManifestProperties struct {
	SubnetID         string
	AvailabilityZone string
	ElasticIP        string
	AccessKeyID      string
	SecretAccessKey  string
	DefaultKeyName   string
	Region           string
	SSLKeyPair       ssl.KeyPair
}

type logger interface {
	Step(message string)
}

type sslKeyPairGenerator interface {
	Generate(string) (ssl.KeyPair, error)
}

func NewManifestBuilder(logger logger, sslKeyPairGenerator sslKeyPairGenerator) ManifestBuilder {
	return ManifestBuilder{
		logger:              logger,
		sslKeyPairGenerator: sslKeyPairGenerator,
	}
}

func (m ManifestBuilder) Build(manifestProperties ManifestProperties) (Manifest, error) {
	m.logger.Step("generating bosh-init manifest")

	releaseManifestBuilder := NewReleaseManifestBuilder()
	resourcePoolsManifestBuilder := NewResourcePoolsManifestBuilder()
	diskPoolsManifestBuilder := NewDiskPoolsManifestBuilder()
	networksManifestBuilder := NewNetworksManifestBuilder()
	jobsManifestBuilder := NewJobsManifestBuilder()
	cloudProviderManifestBuilder := NewCloudProviderManifestBuilder()

	if manifestProperties.SSLKeyPair.IsEmpty() {
		keyPair, err := m.sslKeyPairGenerator.Generate(manifestProperties.ElasticIP)
		if err != nil {
			return Manifest{}, err
		}

		manifestProperties.SSLKeyPair = keyPair
	}

	return Manifest{
		Name:          "bosh",
		Releases:      releaseManifestBuilder.Build(),
		ResourcePools: resourcePoolsManifestBuilder.Build(manifestProperties),
		DiskPools:     diskPoolsManifestBuilder.Build(),
		Networks:      networksManifestBuilder.Build(manifestProperties),
		Jobs:          jobsManifestBuilder.Build(manifestProperties),
		CloudProvider: cloudProviderManifestBuilder.Build(manifestProperties),
	}, nil
}
