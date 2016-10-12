package manifests

type ReleaseManifestBuilder struct{}

func NewReleaseManifestBuilder() ReleaseManifestBuilder {
	return ReleaseManifestBuilder{}
}

func (r ReleaseManifestBuilder) Build(boshURL, boshSHA1, boshAWSCPIURL, boshAWSCPISHA1 string) []Release {
	return []Release{
		{
			Name: "bosh",
			URL:  boshURL,
			SHA1: boshSHA1,
		},
		{
			Name: "bosh-aws-cpi",
			URL:  boshAWSCPIURL,
			SHA1: boshAWSCPISHA1,
		},
	}
}
