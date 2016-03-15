package boshinit

import "fmt"

type CloudProviderManifestBuilder struct {
	uuidGenerator                   UUIDGenerator
	sharedPropertiesManifestBuilder SharedPropertiesManifestBuilder
}

func NewCloudProviderManifestBuilder(uuidGenerator UUIDGenerator) CloudProviderManifestBuilder {
	return CloudProviderManifestBuilder{
		uuidGenerator: uuidGenerator,
	}
}

func (c CloudProviderManifestBuilder) Build(manifestProperties ManifestProperties) (CloudProvider, ManifestProperties, error) {
	sharedPropertiesManifestBuilder := NewSharedPropertiesManifestBuilder()

	password := manifestProperties.Credentials.MBusPassword
	if password == "" {
		var err error
		password, err = c.uuidGenerator.Generate()
		if err != nil {
			return CloudProvider{}, ManifestProperties{}, err
		}

		manifestProperties.Credentials.MBusPassword = password
	}

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

		MBus: fmt.Sprintf("https://mbus:%s@%s:6868", password, manifestProperties.ElasticIP),

		Properties: CloudProviderProperties{
			AWS: sharedPropertiesManifestBuilder.AWS(manifestProperties),

			Agent: AgentProperties{
				MBus: fmt.Sprintf("https://mbus:%s@0.0.0.0:6868", password),
			},

			Blobstore: BlobstoreProperties{
				Provider: "local",
				Path:     "/var/vcap/micro_bosh/data/cache",
			},

			NTP: sharedPropertiesManifestBuilder.NTP(),
		},
	}, manifestProperties, nil
}
