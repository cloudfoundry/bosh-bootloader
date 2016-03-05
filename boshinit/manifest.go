package boshinit

type Manifest struct {
	Name          string         `yaml:"name"`
	Releases      []Release      `yaml:"releases"`
	ResourcePools []ResourcePool `yaml:"resource_pools"`
	DiskPools     []DiskPool     `yaml:"disk_pools"`
	Networks      []Network      `yaml:"networks"`
	Jobs          []Job          `yaml:"jobs"`
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
	InstanceType     string        `yaml:"instance_type"`
	EphemeralDisk    EphemeralDisk `yaml:"ephemeral_disk"`
	AvailabilityZone string        `yaml:"availability_zone"`
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
	Type string `yaml:"type"`
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
	Subnet string `yaml:"subnet"`
}

type Job struct {
	Name               string                   `yaml:"name"`
	Instances          int                      `yaml:"instances"`
	ResourcePool       string                   `yaml:"resource_pool"`
	PersistentDiskPool string                   `yaml:"persistent_disk_pool"`
	Templates          []Template               `yaml:"templates"`
	Networks           []JobNetwork             `yaml:"networks"`
	Properties         map[string]JobProperties `yaml:"properties"`
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

type JobProperties struct {
	Address       string `yaml:"address,omitempty"`
	ListenAddress string `yaml:"listen_address,omitempty"`
	Host          string `yaml:"host,omitempty"`
	User          string `yaml:"user,omitempty"`
	Password      string `yaml:"password,omitempty"`
	Database      string `yaml:"database,omitempty"`
	Adapter       string `yaml:"adapter,omitempty"`
}
