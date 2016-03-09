package boshinit

import "fmt"

type CloudProviderManifestBuilder struct{}

func NewCloudProviderManifestBuilder() CloudProviderManifestBuilder {
	return CloudProviderManifestBuilder{}
}

func (c CloudProviderManifestBuilder) Build(manifestProperties ManifestProperties) CloudProvider {
	sharedPropertiesManifestBuilder := NewSharedPropertiesManifestBuilder()

	return CloudProvider{
		Template: Template{
			Name:    "aws_cpi",
			Release: "bosh-aws-cpi",
		},

		SSHTunnel: SSHTunnel{
			Host:       manifestProperties.ElasticIP,
			Port:       22,
			User:       "vcap",
			PrivateKey: "./bosh.pem",
		},

		MBus: fmt.Sprintf("https://mbus:mbus-password@%s:6868", manifestProperties.ElasticIP),

		Properties: CloudProviderProperties{
			AWS: sharedPropertiesManifestBuilder.AWS(manifestProperties),

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
