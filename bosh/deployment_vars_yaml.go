package bosh

type sharedDeploymentVarsYAML struct {
	AWSYAML          AWSYAML                `yaml:",inline"`
	GCPYAML          GCPYAML                `yaml:",inline"`
	AzureYAML        AzureYAML              `yaml:",inline"`
	VSphereYAML      VSphereYAML            `yaml:",inline"`
	TerraformOutputs map[string]interface{} `yaml:",inline"`
}

type AWSYAML struct {
	AccessKeyID     string `yaml:"access_key_id,omitempty"`
	SecretAccessKey string `yaml:"secret_access_key,omitempty"`
}

type GCPYAML struct {
	Zone           string `yaml:"zone,omitempty"`
	ProjectID      string `yaml:"project_id,omitempty"`
	CredentialJSON string `yaml:"gcp_credentials_json,omitempty"`
}

type AzureYAML struct {
	SubscriptionID string `yaml:"subscription_id,omitempty"`
	TenantID       string `yaml:"tenant_id,omitempty"`
	ClientID       string `yaml:"client_id,omitempty"`
	ClientSecret   string `yaml:"client_secret,omitempty"`
}

type VSphereYAML struct {
	VCenterUser     string `yaml:"vcenter_user,omitempty"`
	VCenterPassword string `yaml:"vcenter_password,omitempty"`
}
