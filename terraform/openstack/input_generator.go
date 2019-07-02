package openstack

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type InputGenerator struct {
}

func NewInputGenerator() InputGenerator {
	return InputGenerator{}
}

func (i InputGenerator) Generate(state storage.State) (map[string]interface{}, error) {
	inputs := map[string]interface{}{
		"env_id":            state.EnvID,
		"auth_url":          state.OpenStack.AuthURL,
		"availability_zone": state.OpenStack.AZ,
		"ext_net_id":        state.OpenStack.NetworkID,
		"ext_net_name":      state.OpenStack.NetworkName,
		"tenant_name":       state.OpenStack.Project,
		"domain_name":       state.OpenStack.Domain,
		"region_name":       state.OpenStack.Region,
	}

	if state.OpenStack.CACertFile != "" {
		inputs["cacert_file"] = state.OpenStack.CACertFile
	}
	if state.OpenStack.Insecure != "" {
		inputs["insecure"] = state.OpenStack.Insecure
	}
	if state.OpenStack.DNSNameServers != nil {
		inputs["dns_nameservers"] = state.OpenStack.DNSNameServers
	}

	return inputs, nil
}

func (i InputGenerator) Credentials(state storage.State) map[string]string {
	return map[string]string{
		"user_name": state.OpenStack.Username,
		"password":  state.OpenStack.Password,
	}
}
