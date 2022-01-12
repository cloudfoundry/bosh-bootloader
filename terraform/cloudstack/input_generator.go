package cloudstack

import (
	"crypto/sha1"
	"fmt"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

const terraformNameCharLimit = 45

type InputGenerator struct {
}

func NewInputGenerator() InputGenerator {
	return InputGenerator{}
}

func (i InputGenerator) Generate(state storage.State) (map[string]interface{}, error) {
	shortEnvID := state.EnvID
	if len(shortEnvID) > terraformNameCharLimit {
		sh1 := fmt.Sprintf("%x", sha1.Sum([]byte(state.EnvID)))
		shortEnvID = fmt.Sprintf("%s-%s", shortEnvID[:terraformNameCharLimit-9], sh1[:8])
	}
	m := map[string]interface{}{
		"env_id":              state.EnvID,
		"cloudstack_endpoint": state.CloudStack.Endpoint,
		"cloudstack_zone":     state.CloudStack.Zone,
		"short_env_id":        shortEnvID,
		"secure":              state.CloudStack.Secure,
		"iso_segment":         state.CloudStack.IsoSegment,
	}
	if state.CloudStack.NetworkVpcOffering != "" {
		m["network_vpc_offering"] = state.CloudStack.NetworkVpcOffering
	}
	if state.CloudStack.ComputeOffering != "" {
		m["cloudstack_compute_offering"] = state.CloudStack.ComputeOffering
	}
	return m, nil
}

func (i InputGenerator) Credentials(state storage.State) map[string]string {
	return map[string]string{
		"cloudstack_api_key":           state.CloudStack.ApiKey,
		"cloudstack_secret_access_key": state.CloudStack.SecretAccessKey,
	}
}
