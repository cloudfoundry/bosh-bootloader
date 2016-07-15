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
	postgresPassword          string
	registryPassword          string
	blobstoreDirectorPassword string
	blobstoreAgentPassword    string
	hmPassword                string
}

func NewJobPropertiesManifestBuilder(natsUsername, postgresUsername, registryUsername, blobstoreDirectorUsername, blobstoreAgentUsername, hmUsername, natsPassword, postgresPassword, registryPassword, blobstoreDirectorPassword, blobstoreAgentPassword, hmPassword string) JobPropertiesManifestBuilder {
	return JobPropertiesManifestBuilder{
		natsUsername:              natsUsername,
		postgresUsername:          postgresUsername,
		registryUsername:          registryUsername,
		blobstoreDirectorUsername: blobstoreDirectorUsername,
		blobstoreAgentUsername:    blobstoreAgentUsername,
		hmUsername:                hmUsername,
		natsPassword:              natsPassword,
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

func (j JobPropertiesManifestBuilder) Postgres() PostgresProperties {
	return PostgresProperties{
		User:     j.postgresUsername,
		Password: j.postgresPassword,
	}
}

func (j JobPropertiesManifestBuilder) Registry() RegistryJobProperties {
	postgres := j.Postgres()
	return RegistryJobProperties{
		Host:     "10.0.0.6",
		Address:  "10.0.0.6",
		Username: j.registryUsername,
		Password: j.registryPassword,
		DB: RegistryPostgresProperties{
			User:     postgres.User,
			Password: postgres.Password,
			Database: "bosh",
		},
		HTTP: HTTPProperties{
			User:     j.registryUsername,
			Password: j.registryPassword,
		},
	}
}

func (j JobPropertiesManifestBuilder) Blobstore() BlobstoreJobProperties {
	return BlobstoreJobProperties{
		Address: "10.0.0.6",
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
		Address:          "127.0.0.1",
		Name:             "my-bosh",
		CPIJob:           "aws_cpi",
		MaxThreads:       10,
		EnablePostDeploy: true,
		DB:               j.Postgres(),
		UserManagement: UserManagementProperties{
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

func (j JobPropertiesManifestBuilder) Agent() AgentProperties {
	return AgentProperties{
		MBus: fmt.Sprintf("nats://%s:%s@10.0.0.6:4222", j.natsUsername, j.natsPassword),
	}
}
