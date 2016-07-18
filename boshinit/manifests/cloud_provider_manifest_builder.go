package manifests

import "fmt"

const MBUS_USERNAME_PREFIX = "mbus-user-"
const MBUS_PASSWORD_PREFIX = "mbus-"

type CloudProviderManifestBuilder struct {
	stringGenerator                 stringGenerator
	sharedPropertiesManifestBuilder SharedPropertiesManifestBuilder
}

func NewCloudProviderManifestBuilder(stringGenerator stringGenerator) CloudProviderManifestBuilder {
	return CloudProviderManifestBuilder{
		stringGenerator: stringGenerator,
	}
}

func (c CloudProviderManifestBuilder) Build(manifestProperties ManifestProperties) (CloudProvider, ManifestProperties, error) {
	sharedPropertiesManifestBuilder := NewSharedPropertiesManifestBuilder()

	username := manifestProperties.Credentials.MBusUsername
	if username == "" {
		var err error
		username, err = c.stringGenerator.Generate(MBUS_USERNAME_PREFIX, PASSWORD_LENGTH)
		if err != nil {
			return CloudProvider{}, ManifestProperties{}, err
		}

		manifestProperties.Credentials.MBusUsername = username
	}

	password := manifestProperties.Credentials.MBusPassword
	if password == "" {
		var err error
		password, err = c.stringGenerator.Generate(MBUS_PASSWORD_PREFIX, PASSWORD_LENGTH)
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

		MBus: fmt.Sprintf("https://%s:%s@%s:6868", username, password, manifestProperties.ElasticIP),

		Properties: CloudProviderProperties{
			AWS: sharedPropertiesManifestBuilder.AWS(manifestProperties),

			Agent: AgentProperties{
				MBus: fmt.Sprintf("https://%s:%s@0.0.0.0:6868", username, password),
			},

			Blobstore: BlobstoreProperties{
				Provider: "local",
				Path:     "/var/vcap/micro_bosh/data/cache",
			},
		},
	}, manifestProperties, nil
}
