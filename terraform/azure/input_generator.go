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
		"env_id":          state.EnvID,
		"simple_env_id":   simpleEnvId,
		"region":        state.Azure.Region,
		"subscription_id": state.Azure.SubscriptionID,
		"tenant_id":       state.Azure.TenantID,
		"client_id":       state.Azure.ClientID,
		"client_secret":   state.Azure.ClientSecret,
	}

	if state.LB.Cert != "" && state.LB.Key != "" {
		input["pfx_cert_base64"] = state.LB.Cert
		input["pfx_password"] = state.LB.Key
		input["system_domain"] = state.LB.Domain
	}

	return input, nil
}
