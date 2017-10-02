package bosh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/cloudfoundry/bosh-bootloader/helpers"
)

const gcpBoshDirectorEphemeralIPOps = `
- type: replace
  path: /networks/name=default/subnets/0/cloud_properties/ephemeral_external_ip?
  value: true
`

const awsBoshDirectorEphemeralIPOps = `
- type: replace
  path: /resource_pools/name=vms/cloud_properties/auto_assign_public_ip?
  value: true
`

const awsEncryptDiskOps = `---
- type: replace
  path: /disk_pools/name=disks/cloud_properties?
  value:
    type: gp2
    encrypted: true
    kms_key_arn: ((kms_key_arn))
`

const azureSSHStaticIP = `
- type: replace
  path: /cloud_provider/ssh_tunnel/host
  value: ((external_ip))
`

const azureJumpboxCpi = `
- type: replace
  path: /releases/-
  value:
    name: bosh-azure-cpi
    url: https://bosh.io/d/github.com/cloudfoundry-incubator/bosh-azure-cpi-release?v=29
    sha1: 630901d22de58597ef8d5a23be9a5b7107d9ecb4

- type: replace
  path: /resource_pools/name=vms/stemcell?
  value:
    url: https://bosh.io/d/stemcells/bosh-azure-hyperv-ubuntu-trusty-go_agent?v=3445.11
    sha1: c70b6854ce1551fbeecfebabfcd6df5215513cad

- type: replace
  path: /resource_pools/name=vms/cloud_properties?
  value:
    instance_type: Standard_D1_v2

- type: replace
  path: /networks/name=private/subnets/0/cloud_properties?
  value:
    resource_group_name: ((resource_group_name))
    virtual_network_name: ((vnet_name))
    subnet_name: ((subnet_name))

- type: replace
  path: /networks/name=public/subnets?/-
  value:
    cloud_properties:
      resource_group_name: ((resource_group_name))

- type: replace
  path: /cloud_provider/template?
  value:
    name: azure_cpi
    release: bosh-azure-cpi

- type: replace
  path: /cloud_provider/ssh_tunnel?
  value:
    host: ((external_ip))
    port: 22
    user: vcap
    private_key: ((private_key))

- type: replace
  path: /cloud_provider/properties/azure?
  value:
    environment: AzureCloud
    subscription_id: ((subscription_id))
    tenant_id: ((tenant_id))
    client_id: ((client_id))
    client_secret: ((client_secret))
    resource_group_name: ((resource_group_name))
    storage_account_name: ((storage_account_name))
    default_security_group: ((default_security_group))
    ssh_user: vcap
    ssh_public_key: ((public_key))
`

type Executor struct {
	command       command
	tempDir       func(string, string) (string, error)
	readFile      func(string) ([]byte, error)
	unmarshalJSON func([]byte, interface{}) error
	marshalJSON   func(interface{}) ([]byte, error)
	writeFile     func(string, []byte, os.FileMode) error
}

type InterpolateInput struct {
	IAAS                   string
	DirectorDeploymentVars string
	JumpboxDeploymentVars  string
	BOSHState              map[string]interface{}
	Variables              string
	OpsFile                string
}

type InterpolateOutput struct {
	Variables string
	Manifest  string
}

type JumpboxInterpolateOutput struct {
	Variables string
	Manifest  string
}

type CreateEnvInput struct {
	Manifest  string
	Variables string
	State     map[string]interface{}
}

type CreateEnvOutput struct {
	State map[string]interface{}
}

type DeleteEnvInput struct {
	Manifest  string
	Variables string
	State     map[string]interface{}
}

type command interface {
	Run(stdout io.Writer, workingDirectory string, args []string) error
}

const VERSION_DEV_BUILD = "[DEV BUILD]"

func NewExecutor(cmd command, tempDir func(string, string) (string, error), readFile func(string) ([]byte, error),
	unmarshalJSON func([]byte, interface{}) error,
	marshalJSON func(interface{}) ([]byte, error), writeFile func(string, []byte, os.FileMode) error) Executor {
	return Executor{
		command:       cmd,
		tempDir:       tempDir,
		readFile:      readFile,
		unmarshalJSON: unmarshalJSON,
		marshalJSON:   marshalJSON,
		writeFile:     writeFile,
	}
}

func (e Executor) JumpboxInterpolate(interpolateInput InterpolateInput) (JumpboxInterpolateOutput, error) {
	tempDir, err := e.tempDir("", "")
	if err != nil {
		return JumpboxInterpolateOutput{}, fmt.Errorf("create temp dir: %s", err)
	}

	var jumpboxSetupFiles = map[string][]byte{
		"jumpbox-deployment-vars.yml": []byte(interpolateInput.JumpboxDeploymentVars),
		"jumpbox.yml":                 MustAsset("vendor/github.com/cppforlife/jumpbox-deployment/jumpbox.yml"),
	}

	if interpolateInput.IAAS == "azure" {
		jumpboxSetupFiles["cpi.yml"] = []byte(azureJumpboxCpi)
	} else {
		jumpboxSetupFiles["cpi.yml"] = MustAsset(filepath.Join("vendor/github.com/cppforlife/jumpbox-deployment", interpolateInput.IAAS, "cpi.yml"))
	}

	if interpolateInput.Variables != "" {
		jumpboxSetupFiles["variables.yml"] = []byte(interpolateInput.Variables)
	}

	for path, contents := range jumpboxSetupFiles {
		err = e.writeFile(filepath.Join(tempDir, path), contents, os.ModePerm)
		if err != nil {
			//not tested
			return JumpboxInterpolateOutput{}, fmt.Errorf("write file: %s", err)
		}
	}

	args := []string{
		"interpolate", filepath.Join(tempDir, "jumpbox.yml"),
		"--var-errs",
		"--vars-store", filepath.Join(tempDir, "variables.yml"),
		"--vars-file", filepath.Join(tempDir, "jumpbox-deployment-vars.yml"),
		"-o", filepath.Join(tempDir, "cpi.yml"),
	}

	buffer := bytes.NewBuffer([]byte{})
	err = e.command.Run(buffer, tempDir, args)
	if err != nil {
		return JumpboxInterpolateOutput{}, fmt.Errorf("bosh interpolate: %s: %s", err, buffer)
	}

	varsStore, err := e.readFile(filepath.Join(tempDir, "variables.yml"))
	if err != nil {
		return JumpboxInterpolateOutput{}, fmt.Errorf("read file: %s", err)
	}

	return JumpboxInterpolateOutput{
		Variables: string(varsStore),
		Manifest:  buffer.String(),
	}, nil
}

func (e Executor) DirectorInterpolate(interpolateInput InterpolateInput) (InterpolateOutput, error) {
	tempDir, err := e.tempDir("", "")
	if err != nil {
		//not tested
		return InterpolateOutput{}, err
	}

	var directorSetupFiles = map[string][]byte{
		"deployment-vars.yml":                    []byte(interpolateInput.DirectorDeploymentVars),
		"user-ops-file.yml":                      []byte(interpolateInput.OpsFile),
		"bosh.yml":                               MustAsset("vendor/github.com/cloudfoundry/bosh-deployment/bosh.yml"),
		"cpi.yml":                                MustAsset(filepath.Join("vendor/github.com/cloudfoundry/bosh-deployment", interpolateInput.IAAS, "cpi.yml")),
		"iam-instance-profile.yml":               MustAsset("vendor/github.com/cloudfoundry/bosh-deployment/aws/iam-instance-profile.yml"),
		"gcp-bosh-director-ephemeral-ip-ops.yml": []byte(gcpBoshDirectorEphemeralIPOps),
		"aws-bosh-director-ephemeral-ip-ops.yml": []byte(awsBoshDirectorEphemeralIPOps),
		"aws-bosh-director-encrypt-disk-ops.yml": []byte(awsEncryptDiskOps),
		"azure-ssh-static-ip.yml":                []byte(azureSSHStaticIP),
		"jumpbox-user.yml":                       MustAsset("vendor/github.com/cloudfoundry/bosh-deployment/jumpbox-user.yml"),
		"gcp-external-ip-not-recommended.yml":    MustAsset("vendor/github.com/cloudfoundry/bosh-deployment/external-ip-not-recommended.yml"),
		"azure-external-ip-not-recommended.yml":  MustAsset("vendor/github.com/cloudfoundry/bosh-deployment/external-ip-not-recommended.yml"),
		"aws-external-ip-not-recommended.yml":    MustAsset("vendor/github.com/cloudfoundry/bosh-deployment/external-ip-with-registry-not-recommended.yml"),
		"uaa.yml":     MustAsset("vendor/github.com/cloudfoundry/bosh-deployment/uaa.yml"),
		"credhub.yml": MustAsset("vendor/github.com/cloudfoundry/bosh-deployment/credhub.yml"),
	}

	if interpolateInput.Variables != "" {
		directorSetupFiles["variables.yml"] = []byte(interpolateInput.Variables)
	}

	for path, contents := range directorSetupFiles {
		err = e.writeFile(filepath.Join(tempDir, path), contents, os.ModePerm)
		if err != nil {
			//not tested
			return InterpolateOutput{}, err
		}
	}

	var args = []string{
		"interpolate", filepath.Join(tempDir, "bosh.yml"),
		"--var-errs",
		"--var-errs-unused",
		"--vars-store", filepath.Join(tempDir, "variables.yml"),
		"--vars-file", filepath.Join(tempDir, "deployment-vars.yml"),
		"-o", filepath.Join(tempDir, "cpi.yml"),
	}

	switch interpolateInput.IAAS {
	case "gcp":
		args = append(args,
			"-o", filepath.Join(tempDir, "jumpbox-user.yml"),
			"-o", filepath.Join(tempDir, "uaa.yml"),
			"-o", filepath.Join(tempDir, "credhub.yml"),
			"-o", filepath.Join(tempDir, "gcp-bosh-director-ephemeral-ip-ops.yml"),
		)
	case "aws":
		args = append(args,
			"-o", filepath.Join(tempDir, "jumpbox-user.yml"),
			"-o", filepath.Join(tempDir, "uaa.yml"),
			"-o", filepath.Join(tempDir, "credhub.yml"),
			"-o", filepath.Join(tempDir, "aws-bosh-director-ephemeral-ip-ops.yml"),
			"-o", filepath.Join(tempDir, "iam-instance-profile.yml"),
			"-o", filepath.Join(tempDir, "aws-bosh-director-encrypt-disk-ops.yml"),
		)
	case "azure":
		args = append(args,
			"-o", filepath.Join(tempDir, "jumpbox-user.yml"),
			"-o", filepath.Join(tempDir, "uaa.yml"),
			"-o", filepath.Join(tempDir, "credhub.yml"),
		)
	}

	buffer := bytes.NewBuffer([]byte{})
	err = e.command.Run(buffer, tempDir, args)
	if err != nil {
		return InterpolateOutput{}, err
	}

	if interpolateInput.OpsFile != "" {
		err = e.writeFile(filepath.Join(tempDir, "bosh.yml"), buffer.Bytes(), os.ModePerm)
		if err != nil {
			//not tested
			return InterpolateOutput{}, err
		}

		args = []string{
			"interpolate", filepath.Join(tempDir, "bosh.yml"),
			"--var-errs",
			"--vars-store", filepath.Join(tempDir, "variables.yml"),
			"--vars-file", filepath.Join(tempDir, "deployment-vars.yml"),
			"-o", filepath.Join(tempDir, "user-ops-file.yml"),
		}

		buffer = bytes.NewBuffer([]byte{})
		err = e.command.Run(buffer, tempDir, args)
		if err != nil {
			return InterpolateOutput{}, err
		}
	}

	varsStore, err := e.readFile(filepath.Join(tempDir, "variables.yml"))
	if err != nil {
		return InterpolateOutput{}, err
	}

	return InterpolateOutput{
		Variables: string(varsStore),
		Manifest:  buffer.String(),
	}, nil
}

func (e Executor) CreateEnv(createEnvInput CreateEnvInput) (CreateEnvOutput, error) {
	tempDir, err := e.writePreviousFiles(createEnvInput.State, createEnvInput.Variables, createEnvInput.Manifest)
	if err != nil {
		return CreateEnvOutput{}, err
	}

	statePath := filepath.Join(tempDir, "state.json")
	variablesPath := filepath.Join(tempDir, "variables.yml")
	manifestPath := filepath.Join(tempDir, "manifest.yml")

	args := []string{
		"create-env", manifestPath,
		"--vars-store", variablesPath,
		"--state", statePath,
	}

	err = e.command.Run(os.Stdout, tempDir, args)
	if err != nil {
		state, readErr := e.readBOSHState(statePath)
		if readErr != nil {
			errorList := helpers.Errors{}
			errorList.Add(err)
			errorList.Add(readErr)
			return CreateEnvOutput{}, errorList
		}

		return CreateEnvOutput{}, NewCreateEnvError(state, err)
	}

	state, err := e.readBOSHState(statePath)
	if err != nil {
		return CreateEnvOutput{}, err
	}

	return CreateEnvOutput{
		State: state,
	}, nil
}

func (e Executor) readBOSHState(statePath string) (map[string]interface{}, error) {
	stateContents, err := e.readFile(statePath)
	if err != nil {
		return map[string]interface{}{}, err
	}

	var state map[string]interface{}
	err = e.unmarshalJSON(stateContents, &state)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return state, nil
}

func (e Executor) DeleteEnv(deleteEnvInput DeleteEnvInput) error {
	tempDir, err := e.writePreviousFiles(deleteEnvInput.State, deleteEnvInput.Variables, deleteEnvInput.Manifest)
	if err != nil {
		return err
	}

	statePath := filepath.Join(tempDir, "state.json")
	variablesPath := filepath.Join(tempDir, "variables.yml")
	boshManifestPath := filepath.Join(tempDir, "manifest.yml")

	args := []string{
		"delete-env", boshManifestPath,
		"--vars-store", variablesPath,
		"--state", statePath,
	}

	err = e.command.Run(os.Stdout, tempDir, args)
	if err != nil {
		state, readErr := e.readBOSHState(statePath)
		if readErr != nil {
			errorList := helpers.Errors{}
			errorList.Add(err)
			errorList.Add(readErr)
			return errorList
		}
		return NewDeleteEnvError(state, err)
	}

	return nil
}

func (e Executor) Version() (string, error) {
	tempDir, err := e.tempDir("", "")
	if err != nil {
		return "", err
	}

	args := []string{"-v"}

	buffer := bytes.NewBuffer([]byte{})
	err = e.command.Run(buffer, tempDir, args)
	if err != nil {
		return "", err
	}

	versionOutput := buffer.String()
	regex := regexp.MustCompile(`\d+.\d+.\d+`)

	version := regex.FindString(versionOutput)
	if version == "" {
		return "", NewBOSHVersionError(errors.New("BOSH version could not be parsed"))
	}

	return version, nil
}

func (e Executor) writePreviousFiles(state map[string]interface{}, variables, manifest string) (string, error) {
	tempDir, err := e.tempDir("", "")
	if err != nil {
		return "", err
	}

	statePath := filepath.Join(tempDir, "state.json")
	variablesPath := filepath.Join(tempDir, "variables.yml")
	boshManifestPath := filepath.Join(tempDir, "manifest.yml")

	if state != nil {
		boshStateContents, err := e.marshalJSON(state)
		if err != nil {
			return "", err
		}
		err = e.writeFile(statePath, boshStateContents, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	err = e.writeFile(variablesPath, []byte(variables), os.ModePerm)
	if err != nil {
		// not tested
		return "", err
	}

	err = e.writeFile(boshManifestPath, []byte(manifest), os.ModePerm)
	if err != nil {
		// not tested
		return "", err
	}

	return tempDir, nil
}
