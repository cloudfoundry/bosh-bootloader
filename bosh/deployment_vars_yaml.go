package bosh

type sharedDeploymentVarsYAML struct {
	TerraformOutputs map[string]interface{} `yaml:",inline"`
}
