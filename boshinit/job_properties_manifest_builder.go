package boshinit

import "fmt"

type JobPropertiesManifestBuilder struct {
	natsPassword              string
	redisPassword             string
	postgresPassword          string
	registryPassword          string
	blobstoreDirectorPassword string
	blobstoreAgentPassword    string
	hmPassword                string
}

func NewJobPropertiesManifestBuilder(natsPassword, redisPassword, postgresPassword, registryPassword, blobstoreDirectorPassword, blobstoreAgentPassword, hmPassword string) JobPropertiesManifestBuilder {
	return JobPropertiesManifestBuilder{
		natsPassword:              natsPassword,
		redisPassword:             redisPassword,
		postgresPassword:          postgresPassword,
		registryPassword:          registryPassword,
		blobstoreDirectorPassword: blobstoreDirectorPassword,
		blobstoreAgentPassword:    blobstoreAgentPassword,
		hmPassword:                hmPassword,
	}
}

func (j JobPropertiesManifestBuilder) NATS() NATSJobProperties {
	return NATSJobProperties{
		Address:  "127.0.0.1",
		User:     "nats",
		Password: j.natsPassword,
	}
}

func (j JobPropertiesManifestBuilder) Redis() RedisJobProperties {
	return RedisJobProperties{
		ListenAddress: "127.0.0.1",
		Address:       "127.0.0.1",
		Password:      j.redisPassword,
	}
}

func (j JobPropertiesManifestBuilder) Postgres() PostgresProperties {
	return PostgresProperties{
		ListenAddress: "127.0.0.1",
		Host:          "127.0.0.1",
		User:          "postgres",
		Password:      j.postgresPassword,
		Database:      "bosh",
		Adapter:       "postgres",
	}
}

func (j JobPropertiesManifestBuilder) Registry() RegistryJobProperties {
	return RegistryJobProperties{
		Address:  "10.0.0.6",
		Host:     "10.0.0.6",
		Username: "admin",
		Password: j.registryPassword,
		Port:     25777,
		DB:       j.Postgres(),
		HTTP: HTTPProperties{
			User:     "admin",
			Password: j.registryPassword,
			Port:     25777,
		},
	}
}

func (j JobPropertiesManifestBuilder) Blobstore() BlobstoreJobProperties {
	return BlobstoreJobProperties{
		Address:  "10.0.0.6",
		Port:     25250,
		Provider: "dav",
		Director: Credentials{
			User:     "director",
			Password: j.blobstoreDirectorPassword,
		},
		Agent: Credentials{
			User:     "agent",
			Password: j.blobstoreAgentPassword,
		},
	}
}

func (j JobPropertiesManifestBuilder) Director(manifestProperties ManifestProperties) DirectorJobProperties {
	return DirectorJobProperties{
		Address:    "127.0.0.1",
		Name:       "my-bosh",
		CPIJob:     "aws_cpi",
		MaxThreads: 10,
		DB:         j.Postgres(),
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
						Password: j.hmPassword,
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

func (j JobPropertiesManifestBuilder) HM() HMJobProperties {
	return HMJobProperties{
		DirectorAccount: Credentials{
			User:     "hm",
			Password: j.hmPassword,
		},
		ResurrectorEnabled: true,
	}
}

func (j JobPropertiesManifestBuilder) Agent() AgentProperties {
	return AgentProperties{
		MBus: fmt.Sprintf("nats://nats:%s@10.0.0.6:4222", j.natsPassword),
	}
}
