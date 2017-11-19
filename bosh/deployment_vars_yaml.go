package bosh

type sharedDeploymentVarsYAML struct {
	InternalCIDR string      `yaml:"internal_cidr,omitempty"`
	InternalGW   string      `yaml:"internal_gw,omitempty"`
	InternalIP   string      `yaml:"internal_ip,omitempty"`
	DirectorName string      `yaml:"director_name,omitempty"`
	ExternalIP   string      `yaml:"external_ip,omitempty"`
	PrivateKey   string      `yaml:"private_key,flow,omitempty"`
	AWSYAML      AWSYAML     `yaml:",inline"`
	GCPYAML      GCPYAML     `yaml:",inline"`
	AzureYAML    AzureYAML   `yaml:",inline"`
	VSphereYAML  VSphereYAML `yaml:",inline"`
}

type AWSYAML struct {
	AZ                    string   `yaml:"az,omitempty"`
	SubnetID              string   `yaml:"subnet_id,omitempty"`
	AccessKeyID           string   `yaml:"access_key_id,omitempty"`
	SecretAccessKey       string   `yaml:"secret_access_key,omitempty"`
	IAMInstanceProfile    string   `yaml:"iam_instance_profile,omitempty"`
	DefaultKeyName        string   `yaml:"default_key_name,omitempty"`
	DefaultSecurityGroups []string `yaml:"default_security_groups,omitempty"`
	Region                string   `yaml:"region,omitempty"`
	KMSKeyARN             string   `yaml:"kms_key_arn,omitempty"`
}

type GCPYAML struct {
	Zone           string   `yaml:"zone,omitempty"`
	Network        string   `yaml:"network,omitempty"`
	Subnetwork     string   `yaml:"subnetwork,omitempty"`
	Tags           []string `json:"tags" yaml:"tags,omitempty"`
	ProjectID      string   `yaml:"project_id,omitempty"`
	CredentialJSON string   `yaml:"gcp_credentials_json,omitempty"`
}

type AzureYAML struct {
	VNetName             string `yaml:"vnet_name,omitempty"`
	SubnetName           string `yaml:"subnet_name,omitempty"`
	SubscriptionID       string `yaml:"subscription_id,omitempty"`
	TenantID             string `yaml:"tenant_id,omitempty"`
	ClientID             string `yaml:"client_id,omitempty"`
	ClientSecret         string `yaml:"client_secret,omitempty"`
	ResourceGroupName    string `yaml:"resource_group_name,omitempty"`
	StorageAccountName   string `yaml:"storage_account_name,omitempty"`
	DefaultSecurityGroup string `yaml:"default_security_group,omitempty"`
	PublicKey            string `yaml:"public_key,flow,omitempty"`
}

type VSphereYAML struct {
	NetworkName      string `yaml:"network_name,omitempty"`
	VCenterRP        string `yaml:"vcenter_rp,omitempty"`
	VCenterCluster   string `yaml:"vcenter_cluster,omitempty"`
	VCenterDC        string `yaml:"vcenter_dc,omitempty"`
	VCenterDisks     string `yaml:"vcenter_disks,omitempty"`
	VCenterDS        string `yaml:"vcenter_ds,omitempty"`
	VCenterIP        string `yaml:"vcenter_ip,omitempty"`
	VCenterVMs       string `yaml:"vcenter_vms,omitempty"`
	VCenterTemplates string `yaml:"vcenter_templates,omitempty"`
	VCenterUser      string `yaml:"vcenter_user,omitempty"`
	VCenterPassword  string `yaml:"vcenter_password,omitempty"`
}
