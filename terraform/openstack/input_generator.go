package openstack

import (
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type InputGenerator struct {
}

func NewInputGenerator() InputGenerator {
	return InputGenerator{}
}

func (i InputGenerator) Generate(state storage.State) (map[string]interface{}, error) {
	cidr := state.OpenStack.InternalCidr
	parsedCIDR, _ := bosh.ParseCIDRBlock(cidr)
	return map[string]interface{}{
		"env_id":                 state.EnvID,
		"internal_cidr":          cidr,
		"internal_gw":            parsedCIDR.GetNthIP(1).String(),
		"director_ip":            parsedCIDR.GetNthIP(6).String(),
		"external_ip":            state.OpenStack.ExternalIP,
		"auth_url":               state.OpenStack.AuthURL,
		"az":                     state.OpenStack.AZ,
		"default_key_name":       state.OpenStack.DefaultKeyName,
		"default_security_group": state.OpenStack.DefaultSecurityGroup,
		"net_id":                 state.OpenStack.NetworkID,
		"openstack_project":      state.OpenStack.Project,
		"openstack_domain":       state.OpenStack.Domain,
		"region":                 state.OpenStack.Region,
		"private_key":            state.OpenStack.PrivateKey,
	}, nil
}

func (i InputGenerator) Credentials(state storage.State) map[string]string {
	return map[string]string{
		"openstack_username": state.OpenStack.Username,
		"openstack_password": state.OpenStack.Password,
	}
}
