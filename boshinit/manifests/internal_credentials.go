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

func NewInternalCredentials(credentials map[string]string) InternalCredentials {
	return InternalCredentials{
		MBusUsername:              credentials["mbusUsername"],
		NatsUsername:              credentials["natsUsername"],
		PostgresUsername:          credentials["postgresUsername"],
		RegistryUsername:          credentials["registryUsername"],
		BlobstoreDirectorUsername: credentials["blobstoreDirectorUsername"],
		BlobstoreAgentUsername:    credentials["blobstoreAgentUsername"],
		HMUsername:                credentials["hmUsername"],
		MBusPassword:              credentials["mbusPassword"],
		NatsPassword:              credentials["natsPassword"],
		RedisPassword:             credentials["redisPassword"],
		PostgresPassword:          credentials["postgresPassword"],
		RegistryPassword:          credentials["registryPassword"],
		BlobstoreDirectorPassword: credentials["blobstoreDirectorPassword"],
		BlobstoreAgentPassword:    credentials["blobstoreAgentPassword"],
		HMPassword:                credentials["hmPassword"],
	}
}

func (c InternalCredentials) ToMap() map[string]string {
	return map[string]string{
		"mbusUsername":              c.MBusUsername,
		"natsUsername":              c.NatsUsername,
		"postgresUsername":          c.PostgresUsername,
		"registryUsername":          c.RegistryUsername,
		"blobstoreDirectorUsername": c.BlobstoreDirectorUsername,
		"blobstoreAgentUsername":    c.BlobstoreAgentUsername,
		"hmUsername":                c.HMUsername,
		"mbusPassword":              c.MBusPassword,
		"natsPassword":              c.NatsPassword,
		"redisPassword":             c.RedisPassword,
		"postgresPassword":          c.PostgresPassword,
		"registryPassword":          c.RegistryPassword,
		"blobstoreDirectorPassword": c.BlobstoreDirectorPassword,
		"blobstoreAgentPassword":    c.BlobstoreAgentPassword,
		"hmPassword":                c.HMPassword,
	}
}
