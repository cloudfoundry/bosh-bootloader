package manifests

type DiskPoolsManifestBuilder struct{}

func NewDiskPoolsManifestBuilder() DiskPoolsManifestBuilder {
	return DiskPoolsManifestBuilder{}
}

func (r DiskPoolsManifestBuilder) Build(iaas string) []DiskPool {
	return []DiskPool{
		{
			Name:     "disks",
			DiskSize: 80 * 1024,
			CloudProperties: DiskPoolsCloudProperties{
				Type:      getDiskType(iaas),
				Encrypted: true,
			},
		},
	}
}

func getDiskType(iaas string) string {
	switch iaas {
	case "aws":
		return "gp2"
	case "gcp":
		return "pd-standard"
	default:
		return ""
	}
}
