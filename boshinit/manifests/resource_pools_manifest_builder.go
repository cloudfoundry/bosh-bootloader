package manifests

type ResourcePoolsManifestBuilder struct{}

func NewResourcePoolsManifestBuilder() ResourcePoolsManifestBuilder {
	return ResourcePoolsManifestBuilder{}
}

func (r ResourcePoolsManifestBuilder) Build(iaas string, manifestProperties ManifestProperties, stemcellURL string, stemcellSHA1 string) []ResourcePool {
	return []ResourcePool{
		{
			Name:    "vms",
			Network: "private",
			Stemcell: Stemcell{
				URL:  stemcellURL,
				SHA1: stemcellSHA1,
			},
			CloudProperties: getCloudProperties(iaas, manifestProperties),
		},
	}
}

func getCloudProperties(iaas string, manifestProperties ManifestProperties) ResourcePoolCloudProperties {
	switch iaas {
	case "aws":
		return ResourcePoolCloudProperties{
			InstanceType: "m3.xlarge",
			EphemeralDisk: EphemeralDisk{
				Size: 25000,
				Type: "gp2",
			},
			AvailabilityZone: manifestProperties.AvailabilityZone,
		}
	case "gcp":
		return ResourcePoolCloudProperties{
			Zone:           manifestProperties.GCP.Zone,
			MachineType:    "n1-standard-4",
			RootDiskSizeGB: 25,
			RootDiskType:   "pd-standard",
			ServiceScopes: []string{
				"compute",
				"devstorage.full_control",
			},
		}
	default:
		return ResourcePoolCloudProperties{}
	}
}
