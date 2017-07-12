package gcp

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
	tfOutputs, err := g.executor.Outputs(tfState)
	if err != nil {
		return map[string]interface{}{}, err
	}

	if val, ok := tfOutputs["system_domain_dns_servers"]; ok {
		servers := []string{}
		for _, server := range val.([]interface{}) {
			servers = append(servers, server.(string))
		}
		tfOutputs["system_domain_dns_servers"] = servers
	}

	return tfOutputs, nil
}
