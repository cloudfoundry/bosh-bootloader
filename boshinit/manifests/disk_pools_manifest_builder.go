package manifests

type DiskPoolsManifestBuilder struct{}

func NewDiskPoolsManifestBuilder() DiskPoolsManifestBuilder {
	return DiskPoolsManifestBuilder{}
}

func (r DiskPoolsManifestBuilder) Build() []DiskPool {
	return []DiskPool{
		{
			Name:     "disks",
			DiskSize: 80 * 1024,
			CloudProperties: DiskPoolsCloudProperties{
				Type:      "gp2",
				Encrypted: true,
			},
		},
	}
}
