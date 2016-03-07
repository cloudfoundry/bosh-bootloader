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

func (JobPropertiesManifestBuilder) Postgres() PostgresJobProperties {
	return PostgresJobProperties{
		ListenAddress: "127.0.0.1",
		Host:          "127.0.0.1",
		User:          "postgres",
		Password:      "postgres-password",
		Database:      "bosh",
		Adapter:       "postgres",
	}
}

func (JobPropertiesManifestBuilder) Registry() RegistryJobProperties {
	return RegistryJobProperties{
		Address:  "10.0.0.6",
		Host:     "10.0.0.6",
		Username: "admin",
		Password: "admin",
		Port:     25777,
		DB: DBProperties{
			ListenAddress: "127.0.0.1",
			Host:          "127.0.0.1",
			User:          "postgres",
			Password:      "postgres-password",
			Database:      "bosh",
			Adapter:       "postgres",
		},
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

func (JobPropertiesManifestBuilder) Director() DirectorJobProperties {
	return DirectorJobProperties{
		Address:    "127.0.0.1",
		Name:       "my-bosh",
		CPIJob:     "aws_cpi",
		MaxThreads: 10,
		DB: DBProperties{
			ListenAddress: "127.0.0.1",
			Host:          "127.0.0.1",
			User:          "postgres",
			Password:      "postgres-password",
			Database:      "bosh",
			Adapter:       "postgres",
		},
		UserManagement: UserManagementProperties{
			Provider: "local",
			Local: LocalProperties{
				Users: []UserProperties{
					{
						Name:     "admin",
						Password: "admin",
					},
					{
						Name:     "hm",
						Password: "hm-password",
					},
				},
			},
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

func (JobPropertiesManifestBuilder) AWS() AWSJobProperties {
	return AWSJobProperties{
		AccessKeyId:           "ACCESS-KEY-ID",
		SecretAccessKey:       "SECRET-ACCESS-KEY",
		DefaultKeyName:        "bosh",
		DefaultSecurityGroups: []string{"bosh"},
		Region:                "REGION",
	}
}

func (JobPropertiesManifestBuilder) Agent() AgentJobProperties {
	return AgentJobProperties{
		MBus: "nats://nats:nats-password@10.0.0.6:4222",
	}
}

func (JobPropertiesManifestBuilder) NTP() []string {
	return []string{"0.pool.ntp.org", "1.pool.ntp.org"}
}
