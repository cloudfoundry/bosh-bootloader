package boshinit

type JobProperties struct {
	NATS      NATSJobProperties      `yaml:"nats"`
	Redis     RedisJobProperties     `yaml:"redis"`
	Postgres  PostgresJobProperties  `yaml:"postgres"`
	Registry  RegistryJobProperties  `yaml:"registry"`
	Blobstore BlobstoreJobProperties `yaml:"blobstore"`
	Director  DirectorJobProperties  `yaml:"director"`
	HM        HMJobProperties        `yaml:"hm"`
	AWS       AWSJobProperties       `yaml:"aws"`
	Agent     AgentJobProperties     `yaml:"agent"`
	NTP       []string               `yaml:"ntp"`
}

type NATSJobProperties struct {
	Address  string `yaml:"address"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type RedisJobProperties struct {
	Address       string `yaml:"address"`
	ListenAddress string `yaml:"listen_address"`
	Password      string `yaml:"password"`
}

type PostgresJobProperties struct {
	ListenAddress string `yaml:"listen_address"`
	Host          string `yaml:"host"`
	User          string `yaml:"user"`
	Password      string `yaml:"password"`
	Database      string `yaml:"database"`
	Adapter       string `yaml:"adapter"`
}

type RegistryJobProperties struct {
	Address  string         `yaml:"address"`
	Host     string         `yaml:"host"`
	Username string         `yaml:"username"`
	Password string         `yaml:"password"`
	Port     int            `yaml:"port"`
	DB       DBProperties   `yaml:"db"`
	HTTP     HTTPProperties `yaml:"http"`
}

type BlobstoreJobProperties struct {
	Address  string      `yaml:"address"`
	Port     int         `yaml:"port"`
	Provider string      `yaml:"provider"`
	Director Credentials `yaml:"director"`
	Agent    Credentials `yaml:"agent"`
}

type DirectorJobProperties struct {
	Address        string                   `yaml:"address"`
	Name           string                   `yaml:"name"`
	CPIJob         string                   `yaml:"cpi_job"`
	MaxThreads     int                      `yaml:"max_threads"`
	DB             DBProperties             `yaml:"db"`
	UserManagement UserManagementProperties `yaml:"user_management"`
}

type AWSJobProperties struct {
	AccessKeyId           string   `yaml:"access_key_id"`
	SecretAccessKey       string   `yaml:"secret_access_key"`
	DefaultKeyName        string   `yaml:"default_key_name"`
	DefaultSecurityGroups []string `yaml:"default_security_groups"`
	Region                string   `yaml:"region"`
}

type AgentJobProperties struct {
	MBus string `yaml:"mbus"`
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
	Provider string          `yaml:"provider"`
	Local    LocalProperties `yaml:"local"`
}

type DBProperties struct {
	ListenAddress string `yaml:"listen_address"`
	Host          string `yaml:"host"`
	User          string `yaml:"user"`
	Password      string `yaml:"password"`
	Database      string `yaml:"database"`
	Adapter       string `yaml:"adapter"`
}

type HTTPProperties struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
}

type Credentials struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}
