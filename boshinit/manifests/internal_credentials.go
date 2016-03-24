package manifests

type InternalCredentials struct {
	MBusUsername              string `json:"mbusUsername"`
	NatsUsername              string `json:"natsUsername"`
	PostgresUsername          string `json:"postgresUsername"`
	RegistryUsername          string `json:"registryUsername"`
	BlobstoreDirectorUsername string `json:"blobstoreDirectorUsername"`
	BlobstoreAgentUsername    string `json:"blobstoreAgentUsername"`
	HMUsername                string `json:"hmUsername"`
	MBusPassword              string `json:"mbusPassword"`
	NatsPassword              string `json:"natsPassword"`
	RedisPassword             string `json:"redisPassword"`
	PostgresPassword          string `json:"postgresPassword"`
	RegistryPassword          string `json:"registryPassword"`
	BlobstoreDirectorPassword string `json:"blobstoreDirectorPassword"`
	BlobstoreAgentPassword    string `json:"blobstoreAgentPassword"`
	HMPassword                string `json:"hmPassword"`
}
