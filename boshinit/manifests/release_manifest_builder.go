package manifests

type ReleaseManifestBuilder struct{}

func NewReleaseManifestBuilder() ReleaseManifestBuilder {
	return ReleaseManifestBuilder{}
}

func (r ReleaseManifestBuilder) Build(boshURL, boshSHA1, boshCPIName, boshCPIURL, boshCPISHA1 string) []Release {
	return []Release{
		{
			Name: "bosh",
			URL:  boshURL,
			SHA1: boshSHA1,
		},
		{
			Name: boshCPIName,
			URL:  boshCPIURL,
			SHA1: boshCPISHA1,
		},
	}
}
