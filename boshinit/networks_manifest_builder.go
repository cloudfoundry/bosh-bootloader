package boshinit

type NetworksManifestBuilder struct{}

func NewNetworksManifestBuilder() NetworksManifestBuilder {
	return NetworksManifestBuilder{}
}

func (r NetworksManifestBuilder) Build(manifestProperties ManifestProperties) []Network {
	return []Network{
		{
			Name: "private",
			Type: "manual",
			Subnets: []Subnet{
				{
					Range:   "10.0.0.0/24",
					Gateway: "10.0.0.1",
					DNS:     []string{"10.0.0.2"},
					CloudProperties: NetworksCloudProperties{
						Subnet: manifestProperties.SubnetID,
					},
				},
			},
		},
		{
			Name: "public",
			Type: "vip",
		},
	}
}
