package stack

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

var (
	stackOutputToTerraformAddr = map[string]string{
		"VPCID":                           "aws_vpc.vpc",
		"VPCInternetGatewayID":            "aws_internet_gateway.ig",
		"NATEIP":                          "aws_eip.nat_eip",
		"NATInstance":                     "aws_instance.nat",
		"NATSecurityGroup":                "aws_security_group.nat_security_group",
		"BOSHEIP":                         "aws_eip.bosh_eip",
		"BOSHSecurityGroup":               "aws_security_group.bosh_security_group",
		"BOSHSubnet":                      "aws_subnet.bosh_subnet",
		"BOSHRouteTable":                  "aws_route_table.bosh_route_table",
		"InternalSecurityGroup":           "aws_security_group.internal_security_group",
		"InternalRouteTable":              "aws_route_table.internal_route_table",
		"CFRouterInternalSecurityGroup":   "aws_security_group.cf_router_lb_internal_security_group",
		"CFRouterSecurityGroup":           "aws_security_group.cf_router_lb_security_group",
		"CFRouterLoadBalancer":            "aws_elb.cf_router_lb",
		"CFSSHProxyInternalSecurityGroup": "aws_security_group.cf_ssh_lb_internal_security_group",
		"CFSSHProxySecurityGroup":         "aws_security_group.cf_ssh_lb_security_group",
		"CFSSHProxyLoadBalancer":          "aws_elb.cf_ssh_lb",
		"ConcourseInternalSecurityGroup":  "aws_security_group.concourse_lb_internal_security_group",
		"ConcourseSecurityGroup":          "aws_security_group.concourse_lb_security_group",
		"ConcourseLoadBalancer":           "aws_elb.concourse_lb",
		"LoadBalancerRouteTable":          "aws_route_table.lb_route_table",
		"LoadBalancerCert":                "aws_iam_server_certificate.lb_cert",
	}
)

//go:generate counterfeiter -o ./fakes/tf.go --fake-name TF . tf
type tf interface {
	Import(terraform.ImportInput) (string, error)
}

//go:generate counterfeiter -o ./fakes/infrastructure.go --fake-name Infrastructure . infrastructure
type infrastructure interface {
	Update(keyPairName string, azs []string, stackName, boshAZ, lbType, lbCertificateARN, envID string) (cloudformation.Stack, error)
	Delete(stackName string) error
}

//go:generate counterfeiter -o ./fakes/zone.go --fake-name Zone . zone
type zone interface {
	Retrieve(region string) ([]string, error)
}

//go:generate counterfeiter -o ./fakes/certificate.go --fake-name Certificate . certificate
type certificate interface {
	Describe(certificateName string) (iam.Certificate, error)
}

//go:generate counterfeiter -o ./fakes/user_policy.go --fake-name UserPolicy . userPolicy
type userPolicy interface {
	Delete(username, policyname string) error
}

//go:generate counterfeiter -o ./fakes/key_pair.go --fake-name KeyPair . keyPair
type keyPair interface {
	Delete(keyPairName string) error
}

type Migrator struct {
	terraform      tf
	infrastructure infrastructure
	certificate    certificate
	userPolicy     userPolicy
	zone           zone
	keyPair        keyPair
}

func NewMigrator(terraform tf,
	infrastructure infrastructure,
	certificate certificate,
	userPolicy userPolicy,
	zone zone,
	keyPair keyPair) Migrator {
	return Migrator{
		terraform:      terraform,
		infrastructure: infrastructure,
		certificate:    certificate,
		userPolicy:     userPolicy,
		zone:           zone,
		keyPair:        keyPair,
	}
}

func (m Migrator) Migrate(state storage.State) (storage.State, error) {
	if state.Stack.Name == "" {
		return state, nil
	}

	availabilityZones, err := m.zone.Retrieve(state.AWS.Region)
	if err != nil {
		return storage.State{}, err
	}

	var (
		certificateARN  string
		certificateName string
	)
	if state.Stack.LBType == "concourse" || state.Stack.LBType == "cf" {
		certificate, err := m.certificate.Describe(state.Stack.CertificateName)
		if err != nil {
			return storage.State{}, err
		}

		certificateARN = certificate.ARN
		certificateName = certificate.Name
		state.LB.Type = state.Stack.LBType
	}

	stack, err := m.infrastructure.Update(state.KeyPair.Name, availabilityZones, state.Stack.Name, state.Stack.BOSHAZ, state.Stack.LBType, certificateARN, state.EnvID)
	if err != nil {
		return storage.State{}, err
	}

	if certificateARN != "" {
		stack.Outputs["LoadBalancerCert"] = certificateName
	}

	var (
		internalSubnetIndex     int
		loadBalancerSubnetIndex int
	)

	for key, value := range stack.Outputs {
		addr := stackOutputToTerraformAddr[key]
		if strings.Contains(key, "InternalSubnet") {
			addr = fmt.Sprintf("aws_subnet.internal_subnets[%d]", internalSubnetIndex)
			internalSubnetIndex++
		}

		if strings.Contains(key, "LoadBalancerSubnet") {
			addr = fmt.Sprintf("aws_subnet.lb_subnets[%d]", loadBalancerSubnetIndex)
			loadBalancerSubnetIndex++
		}

		var err error
		state.TFState, err = m.terraform.Import(terraform.ImportInput{
			TerraformAddr: addr,
			AWSResourceID: value,
			TFState:       state.TFState,
			Creds:         state.AWS,
		})
		if err != nil {
			return storage.State{}, err
		}
	}

	state.MigratedFromCloudFormation = true

	err = m.userPolicy.Delete(fmt.Sprintf("bosh-iam-user-%s", state.EnvID), "aws-cpi")
	if err != nil {
		return storage.State{}, err
	}

	err = m.infrastructure.Delete(state.Stack.Name)
	if err != nil {
		return storage.State{}, err
	}

	err = m.keyPair.Delete(state.KeyPair.Name)
	if err != nil {
		return storage.State{}, err
	}

	state.Stack = storage.Stack{}

	return state, nil
}
