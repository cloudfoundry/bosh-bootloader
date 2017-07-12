package gcp

import "strings"

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

func (g OutputGenerator) Generate(tfState string) (map[string]interface{}, error) {
	outputs, err := g.executor.Outputs(tfState)
	if err != nil {
		return map[string]interface{}{}, err
	}

	if val, ok := outputs["system_domain_dns_servers"]; ok {
		outputs["system_domain_dns_servers"] = strings.Split(val.(string), ",\n")
	}

	return outputs, nil
}
