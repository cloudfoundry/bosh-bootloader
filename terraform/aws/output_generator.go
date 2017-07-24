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

	outputMapping := map[string]string{
		"bosh_eip":                      "external_ip",
		"director_address":              "director_address",
		"bosh_user_access_key":          "bosh_user_access_key",
		"bosh_user_secret_access_key":   "bosh_user_secret_access_key",
		"bosh_subnet_id":                "subnet_id",
		"bosh_subnet_availability_zone": "az",
		"bosh_security_group":           "default_security_groups",
		"internal_security_group":       "internal_security_group",
		"internal_subnet_ids":           "internal_subnet_ids",
		"internal_subnet_cidrs":         "internal_subnet_cidrs",
		"vpc_id":                        "vpc_id",

		"cf_router_lb_name":                    "cf_router_load_balancer",
		"cf_router_lb_url":                     "cf_router_load_balancer_url",
		"cf_router_lb_internal_security_group": "cf_router_internal_security_group",
		"cf_ssh_lb_name":                       "cf_ssh_proxy_load_balancer",
		"cf_ssh_lb_url":                        "cf_ssh_proxy_load_balancer_url",
		"cf_ssh_lb_internal_security_group":    "cf_ssh_proxy_internal_security_group",
		"cf_tcp_lb_name":                       "cf_tcp_router_load_balancer",
		"cf_tcp_lb_url":                        "cf_tcp_router_load_balancer_url",
		"cf_tcp_lb_internal_security_group":    "cf_tcp_router_internal_security_group",

		"concourse_lb_name":                    "concourse_load_balancer",
		"concourse_lb_url":                     "concourse_load_balancer_url",
		"concourse_lb_internal_security_group": "concourse_internal_security_group",

		"env_dns_zone_name_servers": "cf_system_domain_dns_servers",
	}

	if val, ok := tfOutputs["env_dns_zone_name_servers"]; ok {
		servers := []string{}
		for _, server := range val.([]interface{}) {
			servers = append(servers, server.(string))
		}
		tfOutputs["env_dns_zone_name_servers"] = servers
	}

	outputs := map[string]interface{}{}
	for tfKey, outputKey := range outputMapping {
		if value, ok := tfOutputs[tfKey]; ok {
			outputs[outputKey] = value
		}
	}

	return outputs, nil
}
