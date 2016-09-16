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
				URL:  "https://bosh.io/d/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent?v=3262.12",
				SHA1: "90e9825b814da801e1aff7b02508fdada8e155cb",
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
