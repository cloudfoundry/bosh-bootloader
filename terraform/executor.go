package terraform

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

var writeFile func(file string, data []byte, perm os.FileMode) error = ioutil.WriteFile
var readFile func(filename string) ([]byte, error) = ioutil.ReadFile

type Executor struct {
	cmd        terraformCmd
	stateStore stateStore
	debug      bool
}

type ImportInput struct {
	TerraformAddr string
	AWSResourceID string
	TFState       string
	Creds         storage.AWS
}

type tfOutput struct {
	Sensitive bool
	Type      string
	Value     interface{}
}

type terraformCmd interface {
	Run(stdout io.Writer, workingDirectory string, args []string, debug bool) error
}

type stateStore interface {
	GetTerraformDir() (string, error)
	GetVarsDir() (string, error)
}

func NewExecutor(cmd terraformCmd, stateStore stateStore, debug bool) Executor {
	return Executor{
		cmd:        cmd,
		stateStore: stateStore,
		debug:      debug,
	}
}

func (e Executor) Apply(input map[string]string, template, prevTFState string) (string, error) {
	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return "", fmt.Errorf("Get terraform dir: %s", err)
	}

	err = writeFile(filepath.Join(terraformDir, "template.tf"), []byte(template), os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("Write terraform template: %s", err)
	}

	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return "", fmt.Errorf("Get vars dir: %s", err)
	}

	tfStatePath := filepath.Join(varsDir, "terraform.tfstate")

	if prevTFState != "" {
		err = writeFile(tfStatePath, []byte(prevTFState), os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("Write previous terraform state: %s", err)
		}
	}

	err = e.cmd.Run(os.Stdout, terraformDir, []string{"init"}, e.debug)
	if err != nil {
		return "", fmt.Errorf("Run terraform init: %s", err)
	}

	relativeStatePath, err := filepath.Rel(terraformDir, tfStatePath)
	if err != nil {
		return "", fmt.Errorf("Get relative terraform state path: %s", err) //not tested
	}

	args := []string{
		"apply",
		"-state", relativeStatePath,
	}
	for k, v := range input {
		args = append(args, makeVar(k, v)...)
	}
	err = e.cmd.Run(os.Stdout, terraformDir, args, e.debug)
	if err != nil {
		return "", NewExecutorError(tfStatePath, err, e.debug)
	}

	tfState, err := readFile(tfStatePath)
	if err != nil {
		return "", fmt.Errorf("Read terraform state: %s", err)
	}

	return string(tfState), nil
}

func (e Executor) Destroy(input map[string]string, template, prevTFState string) (string, error) {
	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return "", fmt.Errorf("Get terraform dir: %s", err)
	}

	err = writeFile(filepath.Join(terraformDir, "template.tf"), []byte(template), os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("Write terraform template: %s", err)
	}

	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return "", fmt.Errorf("Get vars dir: %s", err)
	}

	tfStatePath := filepath.Join(varsDir, "terraform.tfstate")

	if prevTFState != "" {
		err = writeFile(tfStatePath, []byte(prevTFState), os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("Write previous terraform state: %s", err)
		}
	}

	err = e.cmd.Run(os.Stdout, terraformDir, []string{"init"}, e.debug)
	if err != nil {
		return "", fmt.Errorf("Run terraform init: %s", err)
	}

	relativeStatePath, err := filepath.Rel(terraformDir, tfStatePath)
	if err != nil {
		return "", fmt.Errorf("Get relative terraform state path: %s", err) //not tested
	}

	args := []string{
		"destroy",
		"-force",
		"-state", relativeStatePath,
	}
	for k, v := range input {
		args = append(args, makeVar(k, v)...)
	}
	err = e.cmd.Run(os.Stdout, terraformDir, args, e.debug)
	if err != nil {
		return "", NewExecutorError(tfStatePath, err, e.debug)
	}

	tfState, err := readFile(tfStatePath)
	if err != nil {
		return "", fmt.Errorf("Read terraform state: %s", err)
	}

	return string(tfState), nil
}

func (e Executor) Import(input ImportInput) (string, error) {
	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return "", err
	}

	resourceType := strings.Split(input.TerraformAddr, ".")[0]
	resourceName := strings.Split(input.TerraformAddr, ".")[1]
	resourceName = strings.Split(resourceName, "[")[0]

	template := fmt.Sprintf(`
provider "aws" {
	region     = %q
	access_key = %q
	secret_key = %q
}

resource %q %q {
}`, input.Creds.Region, input.Creds.AccessKeyID, input.Creds.SecretAccessKey, resourceType, resourceName)

	err = writeFile(filepath.Join(terraformDir, "template.tf"), []byte(template), os.ModePerm)
	if err != nil {
		return "", err
	}

	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return "", err
	}

	tfStatePath := filepath.Join(varsDir, "terraform.tfstate")

	err = writeFile(tfStatePath, []byte(input.TFState), os.ModePerm)
	if err != nil {
		return "", err
	}

	err = e.cmd.Run(os.Stdout, terraformDir, []string{"init"}, e.debug)
	if err != nil {
		return "", err
	}

	relativeStatePath, err := filepath.Rel(terraformDir, tfStatePath)
	if err != nil {
		return "", fmt.Errorf("Get relative terraform state path: %s", err) //not tested
	}

	err = e.cmd.Run(os.Stdout, terraformDir, []string{"import", input.TerraformAddr, input.AWSResourceID, "-state", relativeStatePath}, e.debug)
	if err != nil {
		return "", fmt.Errorf("failed to import: %s", err)
	}

	tfStateContents, err := readFile(tfStatePath)
	if err != nil {
		return "", err
	}

	return string(tfStateContents), nil
}

func (e Executor) Version() (string, error) {
	buffer := bytes.NewBuffer([]byte{})
	err := e.cmd.Run(buffer, "/tmp", []string{"version"}, true)
	if err != nil {
		return "", err
	}
	versionOutput := buffer.String()
	regex := regexp.MustCompile(`\d+.\d+.\d+`)

	version := regex.FindString(versionOutput)
	if version == "" {
		return "", errors.New("Terraform version could not be parsed")
	}

	return version, nil
}

func (e Executor) Output(tfState, outputName string) (string, error) {
	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return "", err
	}

	err = writeFile(filepath.Join(terraformDir, "terraform.tfstate"), []byte(tfState), os.ModePerm)
	if err != nil {
		return "", err
	}

	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return "", err
	}

	err = e.cmd.Run(os.Stdout, terraformDir, []string{"init"}, e.debug)
	if err != nil {
		return "", err
	}

	args := []string{"output", outputName, "-state", filepath.Join(varsDir, "terraform.tfstate")}
	buffer := bytes.NewBuffer([]byte{})
	err = e.cmd.Run(buffer, terraformDir, args, true)
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(buffer.String(), "\n"), nil
}

func (e Executor) Outputs(tfState string) (map[string]interface{}, error) {
	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return map[string]interface{}{}, err
	}

	err = writeFile(filepath.Join(varsDir, "terraform.tfstate"), []byte(tfState), os.ModePerm)
	if err != nil {
		return map[string]interface{}{}, err
	}

	err = e.cmd.Run(os.Stdout, varsDir, []string{"init"}, false)
	if err != nil {
		return map[string]interface{}{}, err
	}

	args := []string{"output", "--json"}
	buffer := bytes.NewBuffer([]byte{})
	err = e.cmd.Run(buffer, varsDir, args, true)
	if err != nil {
		return map[string]interface{}{}, err
	}

	var tfOutputs map[string]tfOutput
	err = json.Unmarshal(buffer.Bytes(), &tfOutputs)
	if err != nil {
		return map[string]interface{}{}, err
	}

	outputs := map[string]interface{}{}

	for tfKey, tfValue := range tfOutputs {
		outputs[tfKey] = tfValue.Value
	}

	return outputs, nil
}

func makeVar(name string, value string) []string {
	return []string{"-var", fmt.Sprintf("%s=%s", name, value)}
}
