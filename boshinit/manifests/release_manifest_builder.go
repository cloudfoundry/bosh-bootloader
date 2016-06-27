package manifests

type ReleaseManifestBuilder struct{}

func NewReleaseManifestBuilder() ReleaseManifestBuilder {
	return ReleaseManifestBuilder{}
}

func (r ReleaseManifestBuilder) Build() []Release {
	return []Release{
		{
			Name: "bosh",
			URL:  "https://s3.amazonaws.com/bbl-precompiled-bosh-releases/release-bosh-257.1-on-ubuntu-trusty-stemcell-3262.tgz",
			SHA1: "83393ce5d40590cdb0978292e42027571cb20e17",
		},
		{
			Name: "bosh-aws-cpi",
			URL:  "https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-aws-cpi-release?v=53",
			SHA1: "3a5988bd2b6e951995fe030c75b07c5b922e2d59",
		},
	}
}
