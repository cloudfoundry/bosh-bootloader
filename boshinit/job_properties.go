package boshinit

type JobProperties struct {
	NATS      NATSJobProperties      `yaml:"nats"`
	Redis     RedisJobProperties     `yaml:"redis"`
	Postgres  PostgresProperties     `yaml:"postgres"`
	Registry  RegistryJobProperties  `yaml:"registry"`
	Blobstore BlobstoreJobProperties `yaml:"blobstore"`
	Director  DirectorJobProperties  `yaml:"director"`
	HM        HMJobProperties        `yaml:"hm"`
	AWS       AWSProperties          `yaml:"aws"`
	Agent     AgentProperties        `yaml:"agent"`
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

type RegistryJobProperties struct {
	Address  string             `yaml:"address"`
	Host     string             `yaml:"host"`
	Username string             `yaml:"username"`
	Password string             `yaml:"password"`
	Port     int                `yaml:"port"`
	DB       PostgresProperties `yaml:"db"`
	HTTP     HTTPProperties     `yaml:"http"`
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
	DB             PostgresProperties       `yaml:"db"`
	UserManagement UserManagementProperties `yaml:"user_management"`
	SSL            SSLProperties            `yaml:"ssl"`
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

type SSLProperties struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
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
