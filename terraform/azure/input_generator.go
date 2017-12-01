package azure

import (
	"encoding/base64"
	"io/ioutil"
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

		certBytes, err := ioutil.ReadFile(state.LB.Cert)
		if err != nil {
			return map[string]interface{}{}, err
		}
		certBase64 :=  base64.StdEncoding.EncodeToString(certBytes)
		input["pfx_cert_base64"] = certBase64

		keyBytes, err := ioutil.ReadFile(state.LB.Key)
		if err != nil {
			return map[string]interface{}{}, err
		}
		input["pfx_key"] = string(keyBytes)
	}

	return input, nil
}
