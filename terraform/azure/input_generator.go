package azure

import (
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type InputGenerator struct {
}

func NewInputGenerator() InputGenerator {
	return InputGenerator{}
}

func (i InputGenerator) Generate(state storage.State) (map[string]interface{}, error) {
	simpleEnvId := strings.Replace(state.EnvID, "-", "", -1)
	if len(simpleEnvId) > 20 {
		simpleEnvId = simpleEnvId[:20]
	}
	input := map[string]interface{}{
		"env_id":        state.EnvID,
		"simple_env_id": simpleEnvId,
		"region":        state.Azure.Region,
	}

    if state.Azure.ResourceGroupName != "" {
        input["resource_group_name"] = state.Azure.ResourceGroupName
    }

    if state.Azure.SubnetName != "" && state.Azure.VnetName != "" {
        input["vnet_resource_group_name"] = state.Azure.VnetResourceGroupName
        input["subnet_name"] = state.Azure.SubnetName
        input["vnet_name"] = state.Azure.VnetName
    }

	if state.LB.Cert != "" && state.LB.Key != "" {
		input["pfx_cert_base64"] = state.LB.Cert
		input["pfx_password"] = state.LB.Key
	}

	if state.LB.Domain != "" {
		input["system_domain"] = state.LB.Domain
	}

	return input, nil
}

func (i InputGenerator) Credentials(state storage.State) map[string]string {
	return map[string]string{
		"subscription_id": state.Azure.SubscriptionID,
		"tenant_id":       state.Azure.TenantID,
		"client_id":       state.Azure.ClientID,
		"client_secret":   state.Azure.ClientSecret,
	}
}
