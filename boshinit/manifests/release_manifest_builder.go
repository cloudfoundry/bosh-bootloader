package manifests

type ReleaseManifestBuilder struct{}

func NewReleaseManifestBuilder() ReleaseManifestBuilder {
	return ReleaseManifestBuilder{}
}

func (r ReleaseManifestBuilder) Build() []Release {
	return []Release{
		{
			Name: "bosh",
			URL:  "https://s3.amazonaws.com/bbl-precompiled-bosh-releases/release-bosh-256.2-on-ubuntu-trusty-stemcell-3232.4.tgz",
			SHA1: "bc941575cb8ed25404364fde7c3ff141cecc33eb",
		},
		{
			Name: "bosh-aws-cpi",
			URL:  "https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-aws-cpi-release?v=52",
			SHA1: "dc4a0cca3b33dce291e4fbeb9e9948b6a7be3324",
		},
	}
}
