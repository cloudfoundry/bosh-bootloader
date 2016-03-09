package boshinit

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
				URL:  "https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent?v=3012",
				SHA1: "3380b55948abe4c437dee97f67d2d8df4eec3fc1",
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
