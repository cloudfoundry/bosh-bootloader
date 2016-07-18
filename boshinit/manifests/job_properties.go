package manifests

type JobProperties struct {
	NATS      NATSJobProperties      `yaml:"nats"`
	Postgres  PostgresProperties     `yaml:"postgres"`
	Registry  RegistryJobProperties  `yaml:"registry"`
	Blobstore BlobstoreJobProperties `yaml:"blobstore"`
	Director  DirectorJobProperties  `yaml:"director"`
	HM        HMJobProperties        `yaml:"hm"`
	AWS       AWSProperties          `yaml:"aws"`
	Agent     AgentProperties        `yaml:"agent"`
}

type NATSJobProperties struct {
	Address  string `yaml:"address"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type RegistryJobProperties struct {
	Host     string                     `yaml:"host"`
	Address  string                     `yaml:"address"`
	Username string                     `yaml:"username"`
	Password string                     `yaml:"password"`
	DB       RegistryPostgresProperties `yaml:"db"`
	HTTP     HTTPProperties             `yaml:"http"`
}

type BlobstoreJobProperties struct {
	Address  string      `yaml:"address"`
	Director Credentials `yaml:"director"`
	Agent    Credentials `yaml:"agent"`
}

type DirectorJobProperties struct {
	Address          string                   `yaml:"address"`
	Name             string                   `yaml:"name"`
	CPIJob           string                   `yaml:"cpi_job"`
	MaxThreads       int                      `yaml:"max_threads"`
	EnablePostDeploy bool                     `yaml:"enable_post_deploy"`
	DB               PostgresProperties       `yaml:"db"`
	UserManagement   UserManagementProperties `yaml:"user_management"`
	SSL              SSLProperties            `yaml:"ssl"`
}

type HMJobProperties struct {
	DirectorAccount    Credentials `yaml:"director_account"`
	ResurrectorEnabled bool        `yaml:"resurrector_enabled"`
}

type LocalProperties struct {
	Users []UserProperties `yaml:"users"`
}

type UserProperties struct {
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
}

type UserManagementProperties struct {
	Local LocalProperties `yaml:"local"`
}

type SSLProperties struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

type HTTPProperties struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type Credentials struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}
