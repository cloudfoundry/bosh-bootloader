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
var redactedError = "Some output has been redacted, use `bbl latest-error` to see it or run again with --debug for additional debug output"

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

func (e Executor) Init(template string, input map[string]interface{}) error {
	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return fmt.Errorf("Get terraform dir: %s", err)
	}

	err = writeFile(filepath.Join(terraformDir, "template.tf"), []byte(template), storage.StateMode)
	if err != nil {
		return fmt.Errorf("Write terraform template: %s", err)
	}

	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return fmt.Errorf("Get vars dir: %s", err)
	}

	err = os.MkdirAll(filepath.Join(terraformDir, ".terraform"), os.ModePerm)
	if err != nil {
		return fmt.Errorf("Create .terraform directory: %s", err)
	}

	err = writeFile(filepath.Join(terraformDir, ".terraform", ".gitignore"), []byte("*\n"), storage.StateMode)
	if err != nil {
		return fmt.Errorf("Write .gitignore for terraform binaries: %s", err)
	}

	tfVarsPath := filepath.Join(varsDir, "terraform.tfvars")
	formattedVars := formatVars(input)
	err = writeFile(tfVarsPath, []byte(formattedVars), storage.StateMode)
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
		if vString, ok := value.(string); ok {
			vString = fmt.Sprintf(`"%s"`, vString)
			if strings.Contains(vString, "\n") {
				vString = strings.Replace(vString, "\n", "\\n", -1)
			}
			value = vString
		} else if valList, ok := value.([]string); ok {
			value = fmt.Sprintf(`["%s"]`, strings.Join(valList, `","`))
		}
		formattedVars = fmt.Sprintf("%s\n%s=%s", formattedVars, name, value)
	}
	return formattedVars
}

func (e Executor) runTFCommand(args []string) error {
	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return fmt.Errorf("Get vars dir: %s", err)
	}
	tfStatePath := filepath.Join(varsDir, "terraform.tfstate")

	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return fmt.Errorf("Get terraform dir: %s", err)
	}
	relativeStatePath, err := filepath.Rel(terraformDir, tfStatePath)
	if err != nil {
		return fmt.Errorf("Get relative terraform state path: %s", err) //not tested
	}
	relativeVarsPath, err := filepath.Rel(terraformDir, filepath.Join(varsDir, "terraform.tfvars"))
	if err != nil {
		return fmt.Errorf("Get relative terraform vars path: %s", err) //not tested
	}

	args = append(args,
		"-state", relativeStatePath,
		"-var-file", relativeVarsPath,
	)

	err = e.cmd.Run(os.Stdout, terraformDir, args, e.debug)
	if err != nil {
		if e.debug {
			return err
		} else {
			return fmt.Errorf(redactedError)
		}
	}

	return nil
}

func (e Executor) Apply(credentials map[string]string) error {
	args := []string{"apply", "--auto-approve"}
	for key, value := range credentials {
		arg := fmt.Sprintf("%s=%s", key, value)
		args = append(args, "-var", arg)
	}
	return e.runTFCommand(args)
}

func (e Executor) Destroy(credentials map[string]string) error {
	args := []string{"destroy", "-force"}
	for key, value := range credentials {
		arg := fmt.Sprintf("%s=%s", key, value)
		args = append(args, "-var", arg)
	}
	return e.runTFCommand(args)
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

func (e Executor) Output(outputName string) (string, error) {
	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return "", fmt.Errorf("Get terraform dir: %s", err)
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

func (e Executor) Outputs() (map[string]interface{}, error) {
	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return map[string]interface{}{}, fmt.Errorf("Get vars dir: %s", err)
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
