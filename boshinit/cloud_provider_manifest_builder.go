package boshinit

type CloudProviderManifestBuilder struct{}

func NewCloudProviderManifestBuilder() CloudProviderManifestBuilder {
	return CloudProviderManifestBuilder{}
}

func (c CloudProviderManifestBuilder) Build() CloudProvider {
	sharedPropertiesManifestBuilder := NewSharedPropertiesManifestBuilder()

	return CloudProvider{
		Template: Template{
			Name:    "aws_cpi",
			Release: "bosh-aws-cpi",
		},

		SSHTunnel: SSHTunnel{
			Host:       "ELASTIC-IP",
			Port:       22,
			User:       "vcap",
			PrivateKey: "./bosh.pem",
		},

		MBus: "https://mbus:mbus-password@ELASTIC-IP:6868",

		Properties: CloudProviderProperties{
			AWS: sharedPropertiesManifestBuilder.AWS(),

			Agent: AgentProperties{
				MBus: "https://mbus:mbus-password@0.0.0.0:6868",
			},

			Blobstore: BlobstoreProperties{
				Provider: "local",
				Path:     "/var/vcap/micro_bosh/data/cache",
			},

			NTP: sharedPropertiesManifestBuilder.NTP(),
		},
	}
}
