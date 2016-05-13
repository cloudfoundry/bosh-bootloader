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
			Name:     "default",
			DiskSize: 1024,
			CloudProperties: DiskTypeCloudProperties{
				Type:      "gp2",
				Encrypted: true,
			},
		},
	}
}
