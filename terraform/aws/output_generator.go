package aws

type executor interface {
	Outputs(string) (map[string]interface{}, error)
	Output(string, string) (string, error)
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

	if val, ok := tfOutputs["env_dns_zone_name_servers"]; ok {
		servers := []string{}
		for _, server := range val.([]interface{}) {
			servers = append(servers, server.(string))
		}
		tfOutputs["env_dns_zone_name_servers"] = servers
	}

	return tfOutputs, nil
}
