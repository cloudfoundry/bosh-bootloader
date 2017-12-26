package gcp

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type InputGenerator struct {
}

func NewInputGenerator() InputGenerator {
	return InputGenerator{}
}

func (i InputGenerator) Generate(state storage.State) (map[string]interface{}, error) {
	input := map[string]interface{}{
		"env_id":        state.EnvID,
		"project_id":    state.GCP.ProjectID,
		"region":        state.GCP.Region,
		"zone":          state.GCP.Zone,
		"system_domain": state.LB.Domain,
	}

	if state.LB.Cert != "" && state.LB.Key != "" {
		input["ssl_certificate"] = state.LB.Cert
		input["ssl_certificate_private_key"] = state.LB.Key
	}

	return input, nil
}

func (i InputGenerator) Credentials(state storage.State) map[string]string {
	return map[string]string{
		"credentials": state.GCP.ServiceAccountKeyPath,
	}
}
