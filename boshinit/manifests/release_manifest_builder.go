package manifests

type ReleaseManifestBuilder struct{}

func NewReleaseManifestBuilder() ReleaseManifestBuilder {
	return ReleaseManifestBuilder{}
}

func (r ReleaseManifestBuilder) Build() []Release {
	return []Release{
		{
			Name: "bosh",
			URL:  "https://bosh.io/d/github.com/cloudfoundry/bosh?v=257",
			SHA1: "de801d02d527c686dad63f1fe88cb0b2a959f012",
		},
		{
			Name: "bosh-aws-cpi",
			URL:  "https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-aws-cpi-release?v=53",
			SHA1: "3a5988bd2b6e951995fe030c75b07c5b922e2d59",
		},
	}
}
