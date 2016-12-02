package manifests

type NetworksManifestBuilder struct{}

func NewNetworksManifestBuilder() NetworksManifestBuilder {
	return NetworksManifestBuilder{}
}

func (r NetworksManifestBuilder) Build(manifestProperties ManifestProperties) []Network {
	cloudProperties := NetworksCloudProperties{}

	if manifestProperties.SubnetID != "" {
		cloudProperties = NetworksCloudProperties{
			Subnet: manifestProperties.SubnetID,
		}
	}

	if manifestProperties.GCP.NetworkName != "" {

		ip := false
		cloudProperties = NetworksCloudProperties{
			NetworkName:         manifestProperties.GCP.NetworkName,
			SubnetworkName:      manifestProperties.GCP.SubnetworkName,
			EphemeralExternalIP: &ip,
			Tags: []string{
				manifestProperties.GCP.BOSHTag,
				manifestProperties.GCP.InternalTag,
			},
		}
	}

	return []Network{
		{
			Name: "private",
			Type: "manual",
			Subnets: []Subnet{
				{
					Range:           "10.0.0.0/24",
					Gateway:         "10.0.0.1",
					DNS:             []string{"10.0.0.2"},
					CloudProperties: cloudProperties,
				},
			},
		},
		{
			Name: "public",
			Type: "vip",
		},
	}
}
