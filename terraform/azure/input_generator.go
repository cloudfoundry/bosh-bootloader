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

func (i InputGenerator) Generate(state storage.State) (map[string]string, error) {
	simpleEnvId := strings.Replace(state.EnvID, "-", "", -1)
	if len(simpleEnvId) > 20 {
		simpleEnvId = simpleEnvId[:20]
	}
	input := map[string]string{
		"env_id":          state.EnvID,
		"simple_env_id":   simpleEnvId,
		"subscription_id": state.Azure.SubscriptionID,
		"tenant_id":       state.Azure.TenantID,
		"client_id":       state.Azure.ClientID,
		"client_secret":   state.Azure.ClientSecret,
	}

	return input, nil
}
