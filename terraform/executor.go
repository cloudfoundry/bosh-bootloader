package terraform

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

var redactedError = "Some output has been redacted, use `bbl latest-error` to see it or run again with --debug for additional debug output"

type Executor struct {
	cli          terraformCLI
	bufferingCLI terraformCLI
	stateStore   stateStore
	fs           fs
	debug        bool
	out          io.Writer
}

type tfOutput struct {
	Sensitive bool
	Value     interface{}
}

type terraformCLI interface {
	Run(stdout io.Writer, workingDirectory string, args []string) error
	RunWithEnv(stdout io.Writer, workingDirectory string, args []string, envs []string) error
}

type stateStore interface {
	GetTerraformDir() (string, error)
	GetVarsDir() (string, error)
}

type fs interface {
	fileio.FileWriter
	fileio.DirReader
	fileio.Stater
}

func NewExecutor(cli terraformCLI, bufferingCLI terraformCLI, stateStore stateStore, fs fs, debug bool, out io.Writer) Executor {
	return Executor{
		cli:          cli,
		bufferingCLI: bufferingCLI,
		stateStore:   stateStore,
		fs:           fs,
		debug:        debug,
		out:          out,
	}
}

func (e Executor) Setup(template string, input map[string]interface{}) error {
	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return err
	}

	err = e.fs.WriteFile(filepath.Join(terraformDir, "bbl-template.tf"), []byte(template), storage.StateMode)
	if err != nil {
		return fmt.Errorf("Write terraform template: %s", err)
	}

	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(terraformDir, ".terraform"), os.ModePerm)
	if err != nil {
		return fmt.Errorf("Create .terraform directory: %s", err)
	}

	err = e.fs.WriteFile(filepath.Join(terraformDir, ".terraform", ".gitignore"), []byte("*\n"), storage.StateMode)
	if err != nil {
		return fmt.Errorf("Write .gitignore for terraform binaries: %s", err)
	}

	err = e.fs.WriteFile(filepath.Join(varsDir, "bbl.tfvars"), []byte(formatVars(input)), storage.StateMode)
	if err != nil {
		return fmt.Errorf("Write terraform vars: %s", err)
	}

	return nil
}

func formatVars(inputs map[string]interface{}) string {
	formattedVars := ""
	for name, value := range inputs {
		if vString, ok := value.(string); ok {
			vString = fmt.Sprintf(`"%s"`, vString)
			if strings.Contains(vString, "\n") { //nolint:gosimple
				vString = strings.Replace(vString, "\n", "\\n", -1)
			}
			value = vString
		} else if valList, ok := value.([]string); ok {
			value = fmt.Sprintf(`["%s"]`, strings.Join(valList, `","`))
		}
		formattedVars = fmt.Sprintf("%s\n%s=%v", formattedVars, name, value)
	}
	return formattedVars
}

func (e Executor) runTFCommand(args []string) error {
	return e.runTFCommandWithEnvs(args, []string{})
}

func (e Executor) runTFCommandWithEnvs(args, envs []string) error {
	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return err
	}

	tfStatePath := filepath.Join(varsDir, "terraform.tfstate")

	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return err
	}
	relativeStatePath, err := filepath.Rel(terraformDir, tfStatePath)
	if err != nil {
		return fmt.Errorf("Get relative terraform state path: %s", err) //not tested
	}

	args = append(args,
		"-state", relativeStatePath,
	)

	varsFiles, err := e.fs.ReadDir(varsDir)
	if err != nil {
		return fmt.Errorf("Read contents of vars directory: %s", err)
	}

	for _, file := range varsFiles {
		if strings.HasSuffix(file.Name(), ".tfvars") {
			relativeFilePath, err := filepath.Rel(terraformDir, filepath.Join(varsDir, file.Name()))
			if err != nil {
				return fmt.Errorf("Get relative terraform vars path: %s", err) //not tested
			}
			args = append(args,
				"-var-file", relativeFilePath,
			)
		}
	}

	err = e.cli.RunWithEnv(e.out, terraformDir, args, envs)
	if err != nil {
		if e.debug {
			return err
		}
		return fmt.Errorf("%s", redactedError)
	}

	return nil
}

func (e Executor) Init() error {
	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return err
	}

	err = e.cli.Run(e.out, terraformDir, []string{"init", "--upgrade"})
	if err != nil {
		return fmt.Errorf("Run terraform init --upgrade: %s", err)
	}

	return nil
}

func (e Executor) Apply(credentials map[string]string) error {
	args := []string{"apply"}
	cli, ok := e.cli.(CLI)
	if !ok || !cli.disableTfAutoApprove || !e.debug {
		args = append(args, "--auto-approve")
	}
	if cli.disableTfAutoApprove && !e.debug {
		return fmt.Errorf("%s", "Debug mode is mandatory when terraform auto approve is disabled.")
	}
	for key, value := range credentials {
		arg := fmt.Sprintf("%s=%s", key, value)
		args = append(args, "-var", arg)
	}
	return e.runTFCommand(args)
}

func (e Executor) Validate(credentials map[string]string) error {
	args := []string{"validate"}
	for key, value := range credentials {
		arg := fmt.Sprintf("%s=%s", key, value)
		args = append(args, "-var", arg)
	}

	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return err
	}

	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return err
	}

	varsFiles, err := e.fs.ReadDir(varsDir)
	if err != nil {
		return fmt.Errorf("Read contents of vars directory: %s", err)
	}

	for _, file := range varsFiles {
		if strings.HasSuffix(file.Name(), ".tfvars") {
			relativeFilePath, err := filepath.Rel(terraformDir, filepath.Join(varsDir, file.Name()))
			if err != nil {
				return fmt.Errorf("Get relative terraform vars path: %s", err) //not tested
			}
			args = append(args,
				"-var-file", relativeFilePath,
			)
		}
	}

	err = e.cli.RunWithEnv(e.out, terraformDir, args, []string{})
	if err != nil {
		if e.debug {
			return err
		}
		return fmt.Errorf("%s", redactedError)
	}

	return nil
}

func (e Executor) Destroy(credentials map[string]string) error {
	args := []string{"destroy"}
	cli, ok := e.cli.(CLI)
	if !ok || !cli.disableTfAutoApprove || !e.debug {
		args = append(args, "-auto-approve")
	}
	if cli.disableTfAutoApprove && !e.debug {
		return fmt.Errorf("%s", "Debug mode is mandatory when terraform auto approve is disabled.")
	}
	for key, value := range credentials {
		arg := fmt.Sprintf("%s=%s", key, value)
		args = append(args, "-var", arg)
	}
	return e.runTFCommandWithEnvs(args, []string{"TF_WARN_OUTPUT_ERRORS=1"})
}

func (e Executor) Version() (string, error) {
	buffer := bytes.NewBuffer([]byte{})
	err := e.bufferingCLI.Run(buffer, "/tmp", []string{"version"})
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
		return "", err
	}

	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return "", err
	}

	err = e.cli.Run(e.out, terraformDir, []string{"init"})
	if err != nil {
		return "", fmt.Errorf("Run terraform init in terraform dir: %s", err)
	}

	args := []string{"output", outputName}
	_, err = e.fs.Stat(filepath.Join(varsDir, "terraform.tfstate"))
	if err == nil {
		args = append(args, "-state", filepath.Join(varsDir, "terraform.tfstate"))
	}
	buffer := bytes.NewBuffer([]byte{})
	err = e.bufferingCLI.Run(buffer, terraformDir, args)
	if err != nil {
		return "", fmt.Errorf("Run terraform output -state: %s", err)
	}

	return strings.TrimSuffix(buffer.String(), "\n"), nil
}

func (e Executor) terraformInitIfNeeded(terraformDir string) error {
	_, err := e.fs.Stat(filepath.Join(terraformDir, ".terraform"))
	if err == nil {
		return nil
	}

	return e.Init()
}

func (e Executor) Outputs() (map[string]interface{}, error) {
	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return map[string]interface{}{}, err
	}

	if err = e.terraformInitIfNeeded(terraformDir); err != nil {
		return nil, err
	}

	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return map[string]interface{}{}, err
	}

	buffer := bytes.NewBuffer([]byte{})
	args := []string{"output", "--json"}
	_, err = e.fs.Stat(filepath.Join(varsDir, "terraform.tfstate"))
	if err == nil {
		args = append(args, "-state", filepath.Join(varsDir, "terraform.tfstate"))
	}
	err = e.bufferingCLI.Run(buffer, terraformDir, args)
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

func (e Executor) IsPaved() (bool, error) {
	terraformDir, err := e.stateStore.GetTerraformDir()
	if err != nil {
		return false, err
	}

	err = e.cli.Run(io.Discard, terraformDir, []string{"init"})
	if err != nil {
		return false, fmt.Errorf("Run terraform init in terraform dir: %s", err)
	}

	varsDir, err := e.stateStore.GetVarsDir()
	if err != nil {
		return false, err
	}

	buffer := bytes.NewBuffer([]byte{})
	args := []string{"show"}
	_, err = e.fs.Stat(filepath.Join(varsDir, "terraform.tfstate"))
	if err == nil {
		args = append(args, filepath.Join(varsDir, "terraform.tfstate"))
	}

	err = e.bufferingCLI.Run(buffer, terraformDir, args)
	if err != nil {
		return false, fmt.Errorf("Run terraform show: %s", err)
	}

	if strings.TrimSpace(buffer.String()) == "No state." {
		return false, nil
	}

	return true, nil
}
