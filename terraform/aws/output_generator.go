package aws

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

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

	switch state.LB.Type {
	case "cf":
		outputMapping["cf_router_lb_name"] = "cf_router_load_balancer"
		outputMapping["cf_router_lb_url"] = "cf_router_load_balancer_url"
		outputMapping["cf_router_lb_internal_security_group"] = "cf_router_internal_security_group"
		outputMapping["cf_ssh_lb_name"] = "cf_ssh_proxy_load_balancer"
		outputMapping["cf_ssh_lb_url"] = "cf_ssh_proxy_load_balancer_url"
		outputMapping["cf_ssh_lb_internal_security_group"] = "cf_ssh_proxy_internal_security_group"

		if state.LB.Domain != "" {
			systemDomainDNSServersRaw := tfOutputs["env_dns_zone_name_servers"]
			servers := []string{}
			for _, server := range systemDomainDNSServersRaw.([]interface{}) {
				servers = append(servers, server.(string))
			}
			outputs["cf_system_domain_dns_servers"] = servers
		}
	case "concourse":
		outputMapping["concourse_lb_name"] = "concourse_load_balancer"
		outputMapping["concourse_lb_url"] = "concourse_load_balancer_url"
		outputMapping["concourse_lb_internal_security_group"] = "concourse_internal_security_group"
	}

	for tfKey, outputKey := range outputMapping {
		if value, ok := tfOutputs[tfKey]; ok {
			outputs[outputKey] = value
		}
	}

	return outputs, nil
}
