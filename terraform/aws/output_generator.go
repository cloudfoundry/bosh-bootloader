package aws

import "github.com/cloudfoundry/bosh-bootloader/storage"

type executor interface {
	Outputs(string) (map[string]interface{}, error)
}

type OutputGenerator struct {
	executor executor
}

func NewOutputGenerator(executor executor) OutputGenerator {
	return OutputGenerator{
		executor: executor,
	}
}

func (o OutputGenerator) Generate(state storage.State) (map[string]interface{}, error) {
	tfOutputs, err := o.executor.Outputs(state.TFState)
	if err != nil {
		return map[string]interface{}{}, err
	}

	outputs := map[string]interface{}{}
	outputMapping := map[string]string{
		"bosh_eip":                      "external_ip",
		"bosh_url":                      "director_address",
		"bosh_user_access_key":          "access_key_id",
		"bosh_user_secret_access_key":   "secret_access_key",
		"bosh_subnet_id":                "subnet_id",
		"bosh_subnet_availability_zone": "az",
		"bosh_security_group":           "default_security_groups",
		"internal_security_group":       "internal_security_group",
		"internal_subnet_ids":           "internal_subnet_ids",
		"internal_subnet_cidrs":         "internal_subnet_cidrs",
	}

	for tfKey, outputKey := range outputMapping {
		if value, ok := tfOutputs[tfKey]; ok {
			outputs[outputKey] = value
		}
	}

	return outputs, nil
}
