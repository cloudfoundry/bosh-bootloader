package boshinit

type CloudProviderManifestBuilder struct{}

func NewCloudProviderManifestBuilder() CloudProviderManifestBuilder {
	return CloudProviderManifestBuilder{}
}

func (c CloudProviderManifestBuilder) Build() CloudProvider {
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
			AWS: AWSProperties{
				AccessKeyId:           "ACCESS-KEY-ID",
				SecretAccessKey:       "SECRET-ACCESS-KEY",
				DefaultKeyName:        "bosh",
				DefaultSecurityGroups: []string{"bosh"},
				Region:                "REGION",
			},

			Agent: AgentProperties{
				MBus: "https://mbus:mbus-password@0.0.0.0:6868",
			},

			Blobstore: BlobstoreProperties{
				Provider: "local",
				Path:     "/var/vcap/micro_bosh/data/cache",
			},

			NTP: []string{"0.pool.ntp.org", "1.pool.ntp.org"},
		},
	}
}
