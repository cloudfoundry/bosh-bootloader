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
)

var writeFile func(file string, data []byte, perm os.FileMode) error = ioutil.WriteFile
var readFile func(filename string) ([]byte, error) = ioutil.ReadFile

type Executor struct {
	cmd        terraformCmd
	stateStore stateStore
	debug      bool
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

func (e Executor) IsInitialized() bool {
	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return false // not tested
	}

	_, err = os.Stat(filepath.Join(varsDir, "terraform.tfvars"))
	if err != nil {
		return false
	}

	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return false // not tested
	}

	_, err = os.Stat(filepath.Join(terraformDir, ".terraform"))
	if err != nil {
		return false
	}

	_, err = os.Stat(filepath.Join(terraformDir, "template.tf"))
	if err != nil {
		return false
	}

	return true
}

func (e Executor) Init(template, prevTFState string, input map[string]interface{}) error {
	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return fmt.Errorf("Get terraform dir: %s", err)
	}

	err = writeFile(filepath.Join(terraformDir, "template.tf"), []byte(template), os.ModePerm)
	if err != nil {
		return fmt.Errorf("Write terraform template: %s", err)
	}

	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return fmt.Errorf("Get vars dir: %s", err)
	}

	tfStatePath := filepath.Join(varsDir, "terraform.tfstate")
	if prevTFState != "" {
		err = writeFile(tfStatePath, []byte(prevTFState), os.ModePerm)
		if err != nil {
			return fmt.Errorf("Write previous terraform state: %s", err)
		}
	}

	err = os.MkdirAll(filepath.Join(terraformDir, ".terraform"), os.ModePerm)
	if err != nil {
		return fmt.Errorf("Create .terraform directory: %s", err)
	}

	err = writeFile(filepath.Join(terraformDir, ".terraform", ".gitignore"), []byte("*\n"), os.ModePerm)
	if err != nil {
		return fmt.Errorf("Write .gitignore for terraform binaries: %s", err)
	}

	tfVarsPath := filepath.Join(varsDir, "terraform.tfvars")
	formattedVars := formatVars(input)
	err = writeFile(tfVarsPath, []byte(formattedVars), os.ModePerm)
	if err != nil {
		return fmt.Errorf("Write terraform vars: %s", err)
	}

	err = e.cmd.Run(os.Stdout, terraformDir, []string{"init"}, e.debug)
	if err != nil {
		return fmt.Errorf("Run terraform init: %s", err)
	}

	return nil
}

func formatVars(inputs map[string]interface{}) string {
	formattedVars := ""
	for name, value := range inputs {
		if _, ok := value.(string); ok {
			value = fmt.Sprintf(`"%s"`, value)
		} else if valList, ok := value.([]string); ok {
			value = fmt.Sprintf(`["%s"]`, strings.Join(valList, `","`))
		}
		formattedVars = fmt.Sprintf("%s\n%s=%s", formattedVars, name, value)
	}
	return formattedVars
}

func (e Executor) Apply() (string, error) {
	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return "", fmt.Errorf("Get vars dir: %s", err)
	}
	tfStatePath := filepath.Join(varsDir, "terraform.tfstate")

	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return "", fmt.Errorf("Get terraform dir: %s", err)
	}
	relativeStatePath, err := filepath.Rel(terraformDir, tfStatePath)
	if err != nil {
		return "", fmt.Errorf("Get relative terraform state path: %s", err) //not tested
	}
	relativeVarsPath, err := filepath.Rel(terraformDir, filepath.Join(varsDir, "terraform.tfvars"))
	if err != nil {
		return "", fmt.Errorf("Get relative terraform vars path: %s", err) //not tested
	}

	args := []string{
		"apply",
		"-state", relativeStatePath,
		"-var-file", relativeVarsPath,
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

func (e Executor) Destroy(input map[string]interface{}) (string, error) {
	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return "", fmt.Errorf("Get terraform dir: %s", err)
	}

	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return "", fmt.Errorf("Get vars dir: %s", err)
	}

	tfStatePath := filepath.Join(varsDir, "terraform.tfstate")

	relativeStatePath, err := filepath.Rel(terraformDir, tfStatePath)
	if err != nil {
		return "", fmt.Errorf("Get relative terraform state path: %s", err) //not tested
	}

	relativeVarsPath, err := filepath.Rel(terraformDir, filepath.Join(varsDir, "terraform.tfvars"))
	if err != nil {
		return "", fmt.Errorf("Get relative terraform vars path: %s", err) //not tested
	}

	args := []string{
		"destroy",
		"-force",
		"-state", relativeStatePath,
		"-var-file", relativeVarsPath,
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
		return "", fmt.Errorf("Get terraform dir: %s", err)
	}

	err = writeFile(filepath.Join(terraformDir, "terraform.tfstate"), []byte(tfState), os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("Write terraform state to terraform.tfstate in terraform dir: %s", err)
	}

	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return "", fmt.Errorf("Get vars dir: %s", err)
	}

	err = e.cmd.Run(os.Stdout, terraformDir, []string{"init"}, e.debug)
	if err != nil {
		return "", fmt.Errorf("Run terraform init in terraform dir: %s", err)
	}

	args := []string{"output", outputName, "-state", filepath.Join(varsDir, "terraform.tfstate")}
	buffer := bytes.NewBuffer([]byte{})
	err = e.cmd.Run(buffer, terraformDir, args, true)
	if err != nil {
		return "", fmt.Errorf("Run terraform output -state: %s", err)
	}

	return strings.TrimSuffix(buffer.String(), "\n"), nil
}

func (e Executor) Outputs(tfState string) (map[string]interface{}, error) {
	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("Get vars dir: %s", err)
	}

	err = writeFile(filepath.Join(varsDir, "terraform.tfstate"), []byte(tfState), os.ModePerm)
	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("Write terraform state to terraform.tfstate: %s", err)
	}

	err = e.cmd.Run(os.Stdout, varsDir, []string{"init"}, false)
	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("Run terraform init in vars dir: %s", err)
	}

	buffer := bytes.NewBuffer([]byte{})
	err = e.cmd.Run(buffer, varsDir, []string{"output", "--json"}, true)
	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("Run terraform output --json in vars dir: %s", err)
	}

	tfOutputs := map[string]tfOutput{}
	err = json.Unmarshal(buffer.Bytes(), &tfOutputs)
	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("Unmarshal terraform output: %s", err)
	}

	outputs := map[string]interface{}{}
	for tfKey, tfValue := range tfOutputs {
		outputs[tfKey] = tfValue.Value
	}

	return outputs, nil
}
