package terraform

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/coreos/go-semver/semver"
)

type Manager struct {
	executor              executor
	templateGenerator     templateGenerator
	inputGenerator        inputGenerator
	outputGenerator       outputGenerator
	terraformOutputBuffer *bytes.Buffer
	logger                logger
}

type executor interface {
	Version() (string, error)
	Destroy(inputs map[string]string, terraformTemplate, tfState string) (string, error)
	Apply(inputs map[string]string, terraformTemplate, tfState string) (string, error)
	Import(addr, id, tfState string, creds storage.AWS) (string, error)
}

type templateGenerator interface {
	Generate(storage.State) string
}

type inputGenerator interface {
	Generate(storage.State) (map[string]string, error)
}

type outputGenerator interface {
	Generate(storage.State) (map[string]interface{}, error)
}

type logger interface {
	Step(string, ...interface{})
}

type NewManagerArgs struct {
	Executor              executor
	TemplateGenerator     templateGenerator
	InputGenerator        inputGenerator
	OutputGenerator       outputGenerator
	TerraformOutputBuffer *bytes.Buffer
	Logger                logger
}

func NewManager(args NewManagerArgs) Manager {
	return Manager{
		executor:              args.Executor,
		templateGenerator:     args.TemplateGenerator,
		inputGenerator:        args.InputGenerator,
		outputGenerator:       args.OutputGenerator,
		terraformOutputBuffer: args.TerraformOutputBuffer,
		logger:                args.Logger,
	}
}

func (m Manager) Version() (string, error) {
	return m.executor.Version()
}

func (m Manager) ValidateVersion() error {
	version, err := m.executor.Version()
	if err != nil {
		return err
	}

	currentVersion, err := semver.NewVersion(version)
	if err != nil {
		return err
	}

	// This shouldn't fail, so there is no test for capturing the error.
	minimumVersion, err := semver.NewVersion("0.8.5")
	if err != nil {
		return err
	}

	if currentVersion.LessThan(*minimumVersion) {
		return errors.New("Terraform version must be at least v0.8.5")
	}

	// This shouldn't fail, so there is no test for capturing the error.
	blacklistedVersion, err := semver.NewVersion("0.9.0")
	if err != nil {
		return err
	}

	if currentVersion.Equal(*blacklistedVersion) {
		return errors.New("Version 0.9.0 of terraform is incompatible with bbl, please try a later version.")
	}

	return nil
}

func (m Manager) Apply(bblState storage.State) (storage.State, error) {
	m.logger.Step("generating terraform template")
	template := m.templateGenerator.Generate(bblState)

	input, err := m.inputGenerator.Generate(bblState)
	if err != nil {
		return storage.State{}, err
	}

	tfState, err := m.executor.Apply(
		input,
		template,
		bblState.TFState)

	bblState.LatestTFOutput = readAndReset(m.terraformOutputBuffer)

	switch err.(type) {
	case executorError:
		return storage.State{}, NewManagerError(bblState, err.(executorError))
	case error:
		return storage.State{}, err
	}
	m.logger.Step("applied terraform template")

	bblState.TFState = tfState
	return bblState, nil
}

func (m Manager) Destroy(bblState storage.State) (storage.State, error) {
	m.logger.Step("destroying infrastructure")
	if bblState.TFState == "" {
		return bblState, nil
	}

	template := m.templateGenerator.Generate(bblState)

	input, err := m.inputGenerator.Generate(bblState)
	if err != nil {
		return storage.State{}, err
	}

	tfState, err := m.executor.Destroy(
		input,
		template,
		bblState.TFState)

	bblState.LatestTFOutput = readAndReset(m.terraformOutputBuffer)

	switch err.(type) {
	case executorError:
		return storage.State{}, NewManagerError(bblState, err.(executorError))
	case error:
		return storage.State{}, err
	}
	m.logger.Step("finished destroying infrastructure")

	bblState.TFState = tfState
	return bblState, nil
}

func (m Manager) Import(bblState storage.State, stackOutputs map[string]string) (storage.State, error) {
	var (
		tfState                 string
		internalSubnetIndex     int
		loadBalancerSubnetIndex int
	)

	stackOutputToTerraformAddr := map[string]string{
		"VPCID":                           "aws_vpc.vpc",
		"VPCInternetGatewayID":            "aws_internet_gateway.ig",
		"NATEIP":                          "aws_eip.nat_eip",
		"NATInstance":                     "aws_instance.nat",
		"NATSecurityGroup":                "aws_security_group.nat_security_group",
		"BOSHEIP":                         "aws_eip.bosh_eip",
		"BOSHSecurityGroup":               "aws_security_group.bosh_security_group",
		"BOSHSubnet":                      "aws_subnet.bosh_subnet",
		"InternalSecurityGroup":           "aws_security_group.internal_security_group",
		"CFRouterInternalSecurityGroup":   "aws_security_group.cf_router_lb_internal_security_group",
		"CFRouterSecurityGroup":           "aws_security_group.cf_router_lb_security_group",
		"CFRouterLoadBalancer":            "aws_elb.cf_router_lb",
		"CFSSHProxyInternalSecurityGroup": "aws_security_group.cf_ssh_lb_internal_security_group",
		"CFSSHProxySecurityGroup":         "aws_security_group.cf_ssh_lb_security_group",
		"CFSSHProxyLoadBalancer":          "aws_elb.cf_ssh_lb",
		"ConcourseInternalSecurityGroup":  "aws_security_group.concourse_lb_internal_security_group",
		"ConcourseSecurityGroup":          "aws_security_group.concourse_lb_security_group",
		"ConcourseLoadBalancer":           "aws_elb.concourse_lb",
	}

	for key, value := range stackOutputs {
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
		tfState, err = m.executor.Import(addr, value, tfState, bblState.AWS)
		if err != nil {
			return storage.State{}, err
		}
	}

	bblState.TFState = tfState

	return bblState, nil
}

func (m Manager) GetOutputs(bblState storage.State) (map[string]interface{}, error) {
	outputs, err := m.outputGenerator.Generate(bblState)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return outputs, nil
}

func readAndReset(buf *bytes.Buffer) string {
	contents := buf.Bytes()
	buf.Reset()

	return string(contents)
}
