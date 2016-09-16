package manifests

type ReleaseManifestBuilder struct{}

func NewReleaseManifestBuilder() ReleaseManifestBuilder {
	return ReleaseManifestBuilder{}
}

func (r ReleaseManifestBuilder) Build() []Release {
	return []Release{
		{
			Name: "bosh",
			URL:  "https://s3.amazonaws.com/bbl-precompiled-bosh-releases/release-bosh-257.9-on-ubuntu-trusty-stemcell-3262.12.tgz",
			SHA1: "68125b0e36f599c79f10f8809d328a6dea7e2cd3",
		},
		{
			Name: "bosh-aws-cpi",
			URL:  "https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-aws-cpi-release?v=60",
			SHA1: "8e40a9ff892204007889037f094a1b0d23777058",
		},
	}
}
