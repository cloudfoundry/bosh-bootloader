package boshinit

import (
	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
)

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
	SecurityGroup    string
	SSLKeyPair       ssl.KeyPair
}

type logger interface {
	Step(message string)
	Println(message string)
}

type sslKeyPairGenerator interface {
	Generate(commonName string) (ssl.KeyPair, error)
}

func NewManifestBuilder(logger logger, sslKeyPairGenerator sslKeyPairGenerator) ManifestBuilder {
	return ManifestBuilder{
		logger:              logger,
		sslKeyPairGenerator: sslKeyPairGenerator,
	}
}

func (m ManifestBuilder) Build(manifestProperties ManifestProperties) (Manifest, ManifestProperties, error) {
	m.logger.Step("generating bosh-init manifest")

	releaseManifestBuilder := NewReleaseManifestBuilder()
	resourcePoolsManifestBuilder := NewResourcePoolsManifestBuilder()
	diskPoolsManifestBuilder := NewDiskPoolsManifestBuilder()
	networksManifestBuilder := NewNetworksManifestBuilder()
	jobsManifestBuilder := NewJobsManifestBuilder()
	cloudProviderManifestBuilder := NewCloudProviderManifestBuilder()

	if !manifestProperties.SSLKeyPair.IsValidForIP(manifestProperties.ElasticIP) {
		keyPair, err := m.sslKeyPairGenerator.Generate(manifestProperties.ElasticIP)
		if err != nil {
			return Manifest{}, ManifestProperties{}, err
		}

		manifestProperties.SSLKeyPair = keyPair
	}

	manifest := Manifest{
		Name:          "bosh",
		Releases:      releaseManifestBuilder.Build(),
		ResourcePools: resourcePoolsManifestBuilder.Build(manifestProperties),
		DiskPools:     diskPoolsManifestBuilder.Build(),
		Networks:      networksManifestBuilder.Build(manifestProperties),
		Jobs:          jobsManifestBuilder.Build(manifestProperties),
		CloudProvider: cloudProviderManifestBuilder.Build(manifestProperties),
	}

	yaml, err := candiedyaml.Marshal(manifest)
	if err != nil {
		return Manifest{}, ManifestProperties{}, err
	}

	m.logger.Println(string(yaml))

	return manifest, manifestProperties, nil
}
