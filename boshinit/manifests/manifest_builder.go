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
	Build(string, ManifestProperties) (CloudProvider, ManifestProperties, error)
}

type jobsManifestBuilder interface {
	Build(string, ManifestProperties) ([]Job, ManifestProperties, error)
}

type ManifestProperties struct {
	DirectorName     string
	DirectorUsername string
	DirectorPassword string
	CACommonName     string
	ExternalIP       string
	SSLKeyPair       ssl.KeyPair
	Credentials      InternalCredentials
	AWS              ManifestPropertiesAWS
	GCP              ManifestPropertiesGCP
}

type ManifestPropertiesAWS struct {
	SubnetID         string
	AvailabilityZone string
	AccessKeyID      string
	SecretAccessKey  string
	Region           string
	SecurityGroup    string
	DefaultKeyName   string
}

type ManifestPropertiesGCP struct {
	Zone           string
	NetworkName    string
	SubnetworkName string
	BOSHTag        string
	InternalTag    string
	Project        string
	JsonKey        string
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
	AWSBOSHURL      string
	AWSBOSHSHA1     string
	GCPBOSHURL      string
	GCPBOSHSHA1     string
	BOSHAWSCPIURL   string
	BOSHAWSCPISHA1  string
	BOSHGCPCPIURL   string
	BOSHGCPCPISHA1  string
	AWSStemcellURL  string
	AWSStemcellSHA1 string
	GCPStemcellURL  string
	GCPStemcellSHA1 string
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

func (m ManifestBuilder) Build(iaas string, manifestProperties ManifestProperties) (Manifest, ManifestProperties, error) {
	m.logger.Step("generating bosh-init manifest")

	releaseManifestBuilder := NewReleaseManifestBuilder()
	resourcePoolsManifestBuilder := NewResourcePoolsManifestBuilder()
	diskPoolsManifestBuilder := NewDiskPoolsManifestBuilder()
	networksManifestBuilder := NewNetworksManifestBuilder()

	if !manifestProperties.SSLKeyPair.IsValidForIP(manifestProperties.ExternalIP) {
		keyPair, err := m.sslKeyPairGenerator.Generate(manifestProperties.CACommonName, manifestProperties.ExternalIP)
		if err != nil {
			return Manifest{}, ManifestProperties{}, err
		}

		manifestProperties.SSLKeyPair = keyPair
	}

	cloudProvider, manifestProperties, err := m.cloudProviderManifestBuilder.Build(iaas, manifestProperties)
	if err != nil {
		return Manifest{}, ManifestProperties{}, err
	}

	jobs, manifestProperties, err := m.jobsManifestBuilder.Build(iaas, manifestProperties)
	if err != nil {
		return Manifest{}, ManifestProperties{}, err
	}

	boshURL, boshSHA1 := getBOSHRelease(iaas, m.input.AWSBOSHURL, m.input.AWSBOSHSHA1, m.input.GCPBOSHURL, m.input.GCPBOSHSHA1)
	cpiName, cpiURL, cpiSHA1 := getCPIRelease(iaas, m.input.BOSHAWSCPIURL, m.input.BOSHAWSCPISHA1, m.input.BOSHGCPCPIURL, m.input.BOSHGCPCPISHA1)
	stemcellURL, stemcellSHA1 := getStemcell(iaas, m.input.AWSStemcellURL, m.input.AWSStemcellSHA1, m.input.GCPStemcellURL, m.input.GCPStemcellSHA1)

	return Manifest{
		Name:          "bosh",
		Releases:      releaseManifestBuilder.Build(boshURL, boshSHA1, cpiName, cpiURL, cpiSHA1),
		ResourcePools: resourcePoolsManifestBuilder.Build(iaas, manifestProperties, stemcellURL, stemcellSHA1),
		DiskPools:     diskPoolsManifestBuilder.Build(iaas),
		Networks:      networksManifestBuilder.Build(manifestProperties),
		Jobs:          jobs,
		CloudProvider: cloudProvider,
	}, manifestProperties, nil
}

func getBOSHRelease(iaas, awsURL, awsSHA1, gcpURL, gcpSHA1 string) (url, sha1 string) {
	switch iaas {
	case "aws":
		return awsURL, awsSHA1
	case "gcp":
		return gcpURL, gcpSHA1
	default:
		return "", ""
	}
}

func getCPIRelease(iaas, awsURL, awsSHA1, gcpURL, gcpSHA1 string) (name, url, sha1 string) {
	switch iaas {
	case "aws":
		return "bosh-aws-cpi", awsURL, awsSHA1
	case "gcp":
		return "bosh-google-cpi", gcpURL, gcpSHA1
	default:
		return "", "", ""
	}
}

func getStemcell(iaas, awsURL, awsSHA1, gcpURL, gcpSHA1 string) (url, sha1 string) {
	switch iaas {
	case "aws":
		return awsURL, awsSHA1
	case "gcp":
		return gcpURL, gcpSHA1
	default:
		return "", ""
	}
}
