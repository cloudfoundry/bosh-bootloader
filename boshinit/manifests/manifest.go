package manifests

import "github.com/cloudfoundry/bosh-bootloader/ssl"

const PASSWORD_LENGTH = 15

type Manifest struct {
	Name          string         `yaml:"name"`
	Releases      []Release      `yaml:"releases"`
	ResourcePools []ResourcePool `yaml:"resource_pools"`
	DiskPools     []DiskPool     `yaml:"disk_pools"`
	Networks      []Network      `yaml:"networks"`
	Jobs          []Job          `yaml:"jobs"`
	CloudProvider CloudProvider  `yaml:"cloud_provider"`
}

func (m Manifest) DirectorSSLKeyPair() ssl.KeyPair {
	if len(m.Jobs) < 1 {
		return ssl.KeyPair{}
	}

	return ssl.KeyPair{
		Certificate: []byte(m.Jobs[0].Properties.Director.SSL.Cert),
		PrivateKey:  []byte(m.Jobs[0].Properties.Director.SSL.Key),
	}
}

type Release struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
	SHA1 string `yaml:"sha1"`
}

type ResourcePool struct {
	Name            string                      `yaml:"name"`
	Network         string                      `yaml:"network"`
	Stemcell        Stemcell                    `yaml:"stemcell"`
	CloudProperties ResourcePoolCloudProperties `yaml:"cloud_properties"`
}

type Stemcell struct {
	URL  string `yaml:"url"`
	SHA1 string `yaml:"sha1"`
}

type ResourcePoolCloudProperties struct {
	InstanceType     string        `yaml:"instance_type,omitempty"`
	EphemeralDisk    EphemeralDisk `yaml:"ephemeral_disk,omitempty"`
	AvailabilityZone string        `yaml:"availability_zone,omitempty"`

	Zone           string   `yaml:"zone,omitempty"`
	MachineType    string   `yaml:"machine_type,omitempty"`
	RootDiskSizeGB int      `yaml:"root_disk_size_gb,omitempty"`
	RootDiskType   string   `yaml:"root_disk_type,omitempty"`
	ServiceScopes  []string `yaml:"service_scopes,omitempty"`
}

type EphemeralDisk struct {
	Size int    `yaml:"size"`
	Type string `yaml:"type"`
}

type DiskPool struct {
	Name            string                   `yaml:"name"`
	DiskSize        int                      `yaml:"disk_size"`
	CloudProperties DiskPoolsCloudProperties `yaml:"cloud_properties"`
}

type DiskPoolsCloudProperties struct {
	Type      string `yaml:"type"`
	Encrypted bool   `yaml:"encrypted,omitempty"`
}

type Network struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"`
	Subnets []Subnet `yaml:"subnets,omitempty"`
}

type Subnet struct {
	Range           string                  `yaml:"range"`
	Gateway         string                  `yaml:"gateway"`
	DNS             []string                `yaml:"dns"`
	CloudProperties NetworksCloudProperties `yaml:"cloud_properties"`
}

type NetworksCloudProperties struct {
	Subnet              string   `yaml:"subnet,omitempty"`
	NetworkName         string   `yaml:"network_name,omitempty"`
	SubnetworkName      string   `yaml:"subnetwork_name,omitempty"`
	EphemeralExternalIP *bool    `yaml:"ephemeral_external_ip,omitempty"`
	Tags                []string `yaml:"tags,omitempty"`
}

type Job struct {
	Name               string        `yaml:"name"`
	Instances          int           `yaml:"instances"`
	ResourcePool       string        `yaml:"resource_pool"`
	PersistentDiskPool string        `yaml:"persistent_disk_pool"`
	Templates          []Template    `yaml:"templates"`
	Networks           []JobNetwork  `yaml:"networks"`
	Properties         JobProperties `yaml:"properties"`
}

type Template struct {
	Name    string `yaml:"name"`
	Release string `yaml:"release"`
}

type JobNetwork struct {
	Name      string   `yaml:"name"`
	StaticIPs []string `yaml:"static_ips"`
	Default   []string `yaml:"default,omitempty"`
}

type CloudProvider struct {
	Template   Template                `yaml:"template"`
	SSHTunnel  SSHTunnel               `yaml:"ssh_tunnel"`
	MBus       string                  `yaml:"mbus"`
	Properties CloudProviderProperties `yaml:"properties"`
}

type SSHTunnel struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	User       string `yaml:"user"`
	PrivateKey string `yaml:"private_key"`
}

type CloudProviderProperties struct {
	AWS       AWSProperties       `yaml:"aws,omitempty"`
	Google    GoogleProperties    `yaml:"google,omitempty"`
	Agent     AgentProperties     `yaml:"agent"`
	Blobstore BlobstoreProperties `yaml:"blobstore"`
}

type BlobstoreProperties struct {
	Provider string `yaml:"provider"`
	Path     string `yaml:"path"`
}

type AWSProperties struct {
	AccessKeyId           string   `yaml:"access_key_id"`
	SecretAccessKey       string   `yaml:"secret_access_key"`
	DefaultKeyName        string   `yaml:"default_key_name"`
	DefaultSecurityGroups []string `yaml:"default_security_groups"`
	Region                string   `yaml:"region"`
}

type GoogleProperties struct {
	Project string `yaml:"project"`
	JsonKey string `yaml:"json_key"`
}

type AgentProperties struct {
	MBus string `yaml:"mbus"`
}

type PostgresProperties struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type RegistryPostgresProperties struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}
