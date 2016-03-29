package manifests

type ResourcePoolsManifestBuilder struct{}

func NewResourcePoolsManifestBuilder() ResourcePoolsManifestBuilder {
	return ResourcePoolsManifestBuilder{}
}

func (r ResourcePoolsManifestBuilder) Build(manifestProperties ManifestProperties) []ResourcePool {
	return []ResourcePool{
		{
			Name:    "vms",
			Network: "private",
			Stemcell: Stemcell{
				URL:  "https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent?v=3215",
				SHA1: "84c51fed6342d5eb7cd59728c7d691c75b6c1de8",
			},
			CloudProperties: ResourcePoolCloudProperties{
				InstanceType: "m3.xlarge",
				EphemeralDisk: EphemeralDisk{
					Size: 25000,
					Type: "gp2",
				},
				AvailabilityZone: manifestProperties.AvailabilityZone,
			},
		},
	}
}
