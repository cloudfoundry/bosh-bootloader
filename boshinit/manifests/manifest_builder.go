package manifests

import "github.com/cloudfoundry/bosh-bootloader/ssl"

type logger interface {
	Step(message string, a ...interface{})
	Println(string)
}

type sslKeyPairGenerator interface {
	Generate(caCommonName, commonName string) (ssl.KeyPair, error)
}

type stringGenerator interface {
	Generate(string, int) (string, error)
}

type cloudProviderManifestBuilder interface {
	Build(ManifestProperties) (CloudProvider, ManifestProperties, error)
}

type jobsManifestBuilder interface {
	Build(ManifestProperties) ([]Job, ManifestProperties, error)
}

type ManifestProperties struct {
	DirectorName     string
	DirectorUsername string
	DirectorPassword string
	SubnetID         string
	AvailabilityZone string
	CACommonName     string
	ElasticIP        string
	AccessKeyID      string
	SecretAccessKey  string
	DefaultKeyName   string
	Region           string
	SecurityGroup    string
	SSLKeyPair       ssl.KeyPair
	Credentials      InternalCredentials
}

type ManifestBuilder struct {
	input                        ManifestBuilderInput
	logger                       logger
	sslKeyPairGenerator          sslKeyPairGenerator
	stringGenerator              stringGenerator
	cloudProviderManifestBuilder cloudProviderManifestBuilder
	jobsManifestBuilder          jobsManifestBuilder
}

type ManifestBuilderInput struct {
	BOSHURL        string
	BOSHSHA1       string
	BOSHAWSCPIURL  string
	BOSHAWSCPISHA1 string
	StemcellURL    string
	StemcellSHA1   string
}

func NewManifestBuilder(input ManifestBuilderInput, logger logger, sslKeyPairGenerator sslKeyPairGenerator, stringGenerator stringGenerator, cloudProviderManifestBuilder cloudProviderManifestBuilder, jobsManifestBuilder jobsManifestBuilder) ManifestBuilder {
	return ManifestBuilder{
		input:                        input,
		logger:                       logger,
		sslKeyPairGenerator:          sslKeyPairGenerator,
		stringGenerator:              stringGenerator,
		cloudProviderManifestBuilder: cloudProviderManifestBuilder,
		jobsManifestBuilder:          jobsManifestBuilder,
	}
}

func (m ManifestBuilder) Build(manifestProperties ManifestProperties) (Manifest, ManifestProperties, error) {
	m.logger.Step("generating bosh-init manifest")

	releaseManifestBuilder := NewReleaseManifestBuilder()
	resourcePoolsManifestBuilder := NewResourcePoolsManifestBuilder()
	diskPoolsManifestBuilder := NewDiskPoolsManifestBuilder()
	networksManifestBuilder := NewNetworksManifestBuilder()

	if !manifestProperties.SSLKeyPair.IsValidForIP(manifestProperties.ElasticIP) {
		keyPair, err := m.sslKeyPairGenerator.Generate(manifestProperties.CACommonName, manifestProperties.ElasticIP)
		if err != nil {
			return Manifest{}, ManifestProperties{}, err
		}

		manifestProperties.SSLKeyPair = keyPair
	}

	cloudProvider, manifestProperties, err := m.cloudProviderManifestBuilder.Build(manifestProperties)
	if err != nil {
		return Manifest{}, ManifestProperties{}, err
	}

	jobs, manifestProperties, err := m.jobsManifestBuilder.Build(manifestProperties)
	if err != nil {
		return Manifest{}, ManifestProperties{}, err
	}

	return Manifest{
		Name:          "bosh",
		Releases:      releaseManifestBuilder.Build(m.input.BOSHURL, m.input.BOSHSHA1, m.input.BOSHAWSCPIURL, m.input.BOSHAWSCPISHA1),
		ResourcePools: resourcePoolsManifestBuilder.Build(manifestProperties, m.input.StemcellURL, m.input.StemcellSHA1),
		DiskPools:     diskPoolsManifestBuilder.Build(),
		Networks:      networksManifestBuilder.Build(manifestProperties),
		Jobs:          jobs,
		CloudProvider: cloudProvider,
	}, manifestProperties, nil
}
