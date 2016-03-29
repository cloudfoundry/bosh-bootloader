package manifests

type ReleaseManifestBuilder struct{}

func NewReleaseManifestBuilder() ReleaseManifestBuilder {
	return ReleaseManifestBuilder{}
}

func (r ReleaseManifestBuilder) Build() []Release {
	return []Release{
		{
			Name: "bosh",
			URL:  "https://bosh.io/d/github.com/cloudfoundry/bosh?v=255.6",
			SHA1: "c80989984c4ec4c171f9d880c9f69586dade6389",
		},
		{
			Name: "bosh-aws-cpi",
			URL:  "https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-aws-cpi-release?v=48",
			SHA1: "2abfa1bed326238861e247a10674acf4f7ac48b8",
		},
	}
}
