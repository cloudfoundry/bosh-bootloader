package manifests

type JobProperties struct {
	NATS      NATSJobProperties      `yaml:"nats"`
	Postgres  PostgresProperties     `yaml:"postgres"`
	Registry  RegistryJobProperties  `yaml:"registry"`
	Blobstore BlobstoreJobProperties `yaml:"blobstore"`
	Director  DirectorJobProperties  `yaml:"director"`
	HM        HMJobProperties        `yaml:"hm"`
	AWS       AWSProperties          `yaml:"aws,omitempty"`
	Agent     AgentProperties        `yaml:"agent"`
	Google    GoogleProperties       `yaml:"google,omitempty"`
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

type DefaultSSHOptions struct {
	GatewayHost string `yaml:"gateway_host"`
}

type DirectorJobProperties struct {
	Address                     string                   `yaml:"address"`
	Name                        string                   `yaml:"name"`
	CPIJob                      string                   `yaml:"cpi_job"`
	Workers                     int                      `yaml:"workers"`
	EnableDedicatedStatusWorker bool                     `yaml:"enable_dedicated_status_worker"`
	EnablePostDeploy            bool                     `yaml:"enable_post_deploy"`
	DB                          PostgresProperties       `yaml:"db"`
	UserManagement              UserManagementProperties `yaml:"user_management"`
	SSL                         SSLProperties            `yaml:"ssl"`
	DefaultSSHOptions           DefaultSSHOptions        `yaml:"default_ssh_options"`
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
