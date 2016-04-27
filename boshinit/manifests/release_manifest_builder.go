package manifests

type ReleaseManifestBuilder struct{}

func NewReleaseManifestBuilder() ReleaseManifestBuilder {
	return ReleaseManifestBuilder{}
}

func (r ReleaseManifestBuilder) Build() []Release {
	return []Release{
		{
			Name: "bosh",
			URL:  "https://bosh.io/d/github.com/cloudfoundry/bosh?v=256",
			SHA1: "71701e862c0f4862cb77719d5f3e4f7451da355c",
		},
		{
			Name: "bosh-aws-cpi",
			URL:  "https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-aws-cpi-release?v=51",
			SHA1: "7856e0d1db7d679786fedd3dcb419b802da0434b",
		},
	}
}
