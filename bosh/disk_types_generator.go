package bosh

type DiskTypesGenerator struct{}

type DiskType struct {
	Name            string                  `yaml:"name"`
	DiskSize        int                     `yaml:"disk_size"`
	CloudProperties DiskTypeCloudProperties `yaml:"cloud_properties"`
}

type DiskTypeCloudProperties struct {
	Type      string `yaml:"type"`
	Encrypted bool   `yaml:"encrypted"`
}

func NewDiskTypesGenerator() DiskTypesGenerator {
	return DiskTypesGenerator{}
}

func (DiskTypesGenerator) Generate() []DiskType {
	return []DiskType{
		{
			Name:     "1GB",
			DiskSize: 1024,
			CloudProperties: DiskTypeCloudProperties{
				Type:      "gp2",
				Encrypted: true,
			},
		},
		{
			Name:     "5GB",
			DiskSize: 5120,
			CloudProperties: DiskTypeCloudProperties{
				Type:      "gp2",
				Encrypted: true,
			},
		},
		{
			Name:     "10GB",
			DiskSize: 10240,
			CloudProperties: DiskTypeCloudProperties{
				Type:      "gp2",
				Encrypted: true,
			},
		},
		{
			Name:     "50GB",
			DiskSize: 51200,
			CloudProperties: DiskTypeCloudProperties{
				Type:      "gp2",
				Encrypted: true,
			},
		},
		{
			Name:     "100GB",
			DiskSize: 102400,
			CloudProperties: DiskTypeCloudProperties{
				Type:      "gp2",
				Encrypted: true,
			},
		},
		{
			Name:     "500GB",
			DiskSize: 512000,
			CloudProperties: DiskTypeCloudProperties{
				Type:      "gp2",
				Encrypted: true,
			},
		},
		{
			Name:     "1TB",
			DiskSize: 1048576,
			CloudProperties: DiskTypeCloudProperties{
				Type:      "gp2",
				Encrypted: true,
			},
		},
	}
}
