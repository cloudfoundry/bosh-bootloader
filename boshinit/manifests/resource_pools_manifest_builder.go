package manifests

type ResourcePoolsManifestBuilder struct{}

func NewResourcePoolsManifestBuilder() ResourcePoolsManifestBuilder {
	return ResourcePoolsManifestBuilder{}
}

func (r ResourcePoolsManifestBuilder) Build(manifestProperties ManifestProperties, stemcellURL string, stemcellSHA1 string) []ResourcePool {
	return []ResourcePool{
		{
			Name:    "vms",
			Network: "private",
			Stemcell: Stemcell{
				URL:  stemcellURL,
				SHA1: stemcellSHA1,
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
