package manifests

type ReleaseManifestBuilder struct{}

func NewReleaseManifestBuilder() ReleaseManifestBuilder {
	return ReleaseManifestBuilder{}
}

func (r ReleaseManifestBuilder) Build() []Release {
	return []Release{
		{
			Name: "bosh",
			URL:  "https://bosh.io/d/github.com/cloudfoundry/bosh?v=255.3",
			SHA1: "1a3d61f968b9719d9afbd160a02930c464958bf4",
		},
		{
			Name: "bosh-aws-cpi",
			URL:  "https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-aws-cpi-release?v=44",
			SHA1: "a1fe03071e8b9bf1fa97a4022151081bf144c8bc",
		},
	}
}
