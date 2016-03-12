package boshinit

type JobPropertiesManifestBuilder struct{}

func NewJobPropertiesManifestBuilder() JobPropertiesManifestBuilder {
	return JobPropertiesManifestBuilder{}
}

func (JobPropertiesManifestBuilder) NATS() NATSJobProperties {
	return NATSJobProperties{
		Address:  "127.0.0.1",
		User:     "nats",
		Password: "nats-password",
	}
}

func (JobPropertiesManifestBuilder) Redis() RedisJobProperties {
	return RedisJobProperties{
		ListenAddress: "127.0.0.1",
		Address:       "127.0.0.1",
		Password:      "redis-password",
	}
}

func (JobPropertiesManifestBuilder) Registry() RegistryJobProperties {
	sharedPropertiesManifestBuilder := NewSharedPropertiesManifestBuilder()
	return RegistryJobProperties{
		Address:  "10.0.0.6",
		Host:     "10.0.0.6",
		Username: "admin",
		Password: "admin",
		Port:     25777,
		DB:       sharedPropertiesManifestBuilder.Postgres(),
		HTTP: HTTPProperties{
			User:     "admin",
			Password: "admin",
			Port:     25777,
		},
	}
}

func (JobPropertiesManifestBuilder) Blobstore() BlobstoreJobProperties {
	return BlobstoreJobProperties{
		Address:  "10.0.0.6",
		Port:     25250,
		Provider: "dav",
		Director: Credentials{
			User:     "director",
			Password: "director-password",
		},
		Agent: Credentials{
			User:     "agent",
			Password: "agent-password",
		},
	}
}

func (j JobPropertiesManifestBuilder) Director(manifestProperties ManifestProperties) DirectorJobProperties {
	sharedPropertiesManifestBuilder := NewSharedPropertiesManifestBuilder()
	return DirectorJobProperties{
		Address:    "127.0.0.1",
		Name:       "my-bosh",
		CPIJob:     "aws_cpi",
		MaxThreads: 10,
		DB:         sharedPropertiesManifestBuilder.Postgres(),
		UserManagement: UserManagementProperties{
			Provider: "local",
			Local: LocalProperties{
				Users: []UserProperties{
					{
						Name:     manifestProperties.DirectorUsername,
						Password: manifestProperties.DirectorPassword,
					},
					{
						Name:     "hm",
						Password: "hm-password",
					},
				},
			},
		},
		SSL: SSLProperties{
			Cert: string(manifestProperties.SSLKeyPair.Certificate),
			Key:  string(manifestProperties.SSLKeyPair.PrivateKey),
		},
	}
}

func (JobPropertiesManifestBuilder) HM() HMJobProperties {
	return HMJobProperties{
		DirectorAccount: Credentials{
			User:     "hm",
			Password: "hm-password",
		},
		ResurrectorEnabled: true,
	}
}

func (JobPropertiesManifestBuilder) Agent() AgentProperties {
	return AgentProperties{
		MBus: "nats://nats:nats-password@10.0.0.6:4222",
	}
}
