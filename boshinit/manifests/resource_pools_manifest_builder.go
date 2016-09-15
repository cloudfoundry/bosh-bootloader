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
				URL:  "https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent?v=3262",
				SHA1: "f04361747243dadc6e13ce74f5044b46931fb00a",
			},
			CloudProperties: ResourcePoolCloudProperties{
				InstanceType: "m3.xlarge",
				EphemeralDisk: EphemeralDisk{
					Size: 80 * 1024,
					Type: "gp2",
				},
				AvailabilityZone: manifestProperties.AvailabilityZone,
			},
		},
	}
}
