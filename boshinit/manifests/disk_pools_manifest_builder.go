package manifests

type DiskPoolsManifestBuilder struct{}

func NewDiskPoolsManifestBuilder() DiskPoolsManifestBuilder {
	return DiskPoolsManifestBuilder{}
}

func (r DiskPoolsManifestBuilder) Build() []DiskPool {
	return []DiskPool{
		{
			Name:     "disks",
			DiskSize: 20000,
			CloudProperties: DiskPoolsCloudProperties{
				Type:      "gp2",
				Encrypted: true,
			},
		},
	}
}
