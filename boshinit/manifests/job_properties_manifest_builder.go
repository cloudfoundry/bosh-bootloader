package manifests

import "fmt"

type JobPropertiesManifestBuilder struct {
	natsUsername              string
	postgresUsername          string
	registryUsername          string
	blobstoreDirectorUsername string
	blobstoreAgentUsername    string
	hmUsername                string
	natsPassword              string
	redisPassword             string
	postgresPassword          string
	registryPassword          string
	blobstoreDirectorPassword string
	blobstoreAgentPassword    string
	hmPassword                string
}

func NewJobPropertiesManifestBuilder(natsUsername, postgresUsername, registryUsername, blobstoreDirectorUsername, blobstoreAgentUsername, hmUsername, natsPassword, redisPassword, postgresPassword, registryPassword, blobstoreDirectorPassword, blobstoreAgentPassword, hmPassword string) JobPropertiesManifestBuilder {
	return JobPropertiesManifestBuilder{
		natsUsername:              natsUsername,
		postgresUsername:          postgresUsername,
		registryUsername:          registryUsername,
		blobstoreDirectorUsername: blobstoreDirectorUsername,
		blobstoreAgentUsername:    blobstoreAgentUsername,
		hmUsername:                hmUsername,
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
		User:     j.natsUsername,
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
		User:          j.postgresUsername,
		Password:      j.postgresPassword,
		Database:      "bosh",
		Adapter:       "postgres",
	}
}

func (j JobPropertiesManifestBuilder) Registry() RegistryJobProperties {
	return RegistryJobProperties{
		Address:  "10.0.0.6",
		Host:     "10.0.0.6",
		Username: j.registryUsername,
		Password: j.registryPassword,
		Port:     25777,
		DB:       j.Postgres(),
		HTTP: HTTPProperties{
			User:     j.registryUsername,
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
			User:     j.blobstoreDirectorUsername,
			Password: j.blobstoreDirectorPassword,
		},
		Agent: Credentials{
			User:     j.blobstoreAgentUsername,
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
						Name:     j.hmUsername,
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
			User:     j.hmUsername,
			Password: j.hmPassword,
		},
		ResurrectorEnabled: true,
	}
}

func (j JobPropertiesManifestBuilder) DNS() DNSJobProperties {
	return DNSJobProperties{
		Address:    "10.0.0.6",
		DB:         j.Postgres(),
	}
}

func (j JobPropertiesManifestBuilder) Agent() AgentProperties {
	return AgentProperties{
		MBus: fmt.Sprintf("nats://%s:%s@10.0.0.6:4222", j.natsUsername, j.natsPassword),
	}
}
