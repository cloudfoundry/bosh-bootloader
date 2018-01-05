package bosh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Executor struct {
	command       command
	readFile      func(string) ([]byte, error)
	unmarshalJSON func([]byte, interface{}) error
	marshalJSON   func(interface{}) ([]byte, error)
	writeFile     func(string, []byte, os.FileMode) error
}

type InterpolateInput struct {
	DeploymentDir string
	StateDir      string
	VarsDir       string
	IAAS          string
}

type CreateEnvInput struct {
	StateDir       string
	VarsDir        string
	Deployment     string
	DeploymentVars string
}

type DeleteEnvInput struct {
	StateDir   string
	VarsDir    string
	Deployment string
}

type command interface {
	GetBOSHPath() (string, error)
	Run(stdout io.Writer, workingDirectory string, args []string) error
}

type setupFile struct {
	source   string
	dest     string
	contents []byte
}

var (
	jumpboxDeploymentRepo = "vendor/github.com/cppforlife/jumpbox-deployment"
	boshDeploymentRepo    = "vendor/github.com/cloudfoundry/bosh-deployment"
)

func NewExecutor(cmd command, readFile func(string) ([]byte, error),
	unmarshalJSON func([]byte, interface{}) error,
	marshalJSON func(interface{}) ([]byte, error),
	writeFile func(string, []byte, os.FileMode) error) Executor {
	return Executor{
		command:       cmd,
		readFile:      readFile,
		unmarshalJSON: unmarshalJSON,
		marshalJSON:   marshalJSON,
		writeFile:     writeFile,
	}
}

func (e Executor) getSetupFiles(sourcePath, destPath string) []setupFile {
	files := []setupFile{}

	assetNames := AssetNames()
	for _, asset := range assetNames {
		if strings.Contains(asset, sourcePath) {
			files = append(files, setupFile{
				source:   strings.TrimPrefix(asset, sourcePath),
				dest:     filepath.Join(destPath, strings.TrimPrefix(asset, sourcePath)),
				contents: MustAsset(asset),
			})
		}
	}
	return files
}

func (e Executor) PlanJumpbox(input InterpolateInput) error {
	setupFiles := e.getSetupFiles(jumpboxDeploymentRepo, input.DeploymentDir)

	for _, f := range setupFiles {
		os.MkdirAll(filepath.Dir(f.dest), os.ModePerm)
		err := e.writeFile(f.dest, f.contents, storage.StateMode)
		if err != nil {
			return fmt.Errorf("Jumpbox write setup file: %s", err) //not tested
		}
	}

	sharedArgs := []string{
		"--vars-store", filepath.Join(input.VarsDir, "jumpbox-vars-store.yml"),
		"--vars-file", filepath.Join(input.VarsDir, "jumpbox-vars-file.yml"),
		"-o", filepath.Join(input.DeploymentDir, input.IAAS, "cpi.yml"),
	}

	if input.IAAS == "vsphere" {
		sharedArgs = append(sharedArgs, "-o", filepath.Join(input.DeploymentDir, "vsphere", "resource-pool.yml"))
		vSphereJumpboxNetworkOpsPath := filepath.Join(input.DeploymentDir, "vsphere-jumpbox-network.yml")
		sharedArgs = append(sharedArgs, "-o", vSphereJumpboxNetworkOpsPath)
		err := e.writeFile(vSphereJumpboxNetworkOpsPath, []byte(VSphereJumpboxNetworkOps), os.ModePerm)
		if err != nil {
			return fmt.Errorf("Jumpbox write vsphere network ops file: %s", err) //not tested
		}
	}

	jumpboxState := filepath.Join(input.VarsDir, "jumpbox-state.json")

	boshArgs := append([]string{
		filepath.Join(input.DeploymentDir, "jumpbox.yml"),
		"--state", jumpboxState,
	}, sharedArgs...)

	boshPath, err := e.command.GetBOSHPath()
	if err != nil {
		return fmt.Errorf("Jumpbox get BOSH path: %s", err) //not tested
	}

	createEnvCmd := []byte(formatScript(boshPath, input.StateDir, "create-env", boshArgs))
	createJumpboxScript := filepath.Join(input.StateDir, "create-jumpbox.sh")
	err = e.writeFile(createJumpboxScript, createEnvCmd, 0750)
	if err != nil {
		return err
	}

	deleteEnvCmd := []byte(formatScript(boshPath, input.StateDir, "delete-env", boshArgs))
	deleteJumpboxScript := filepath.Join(input.StateDir, "delete-jumpbox.sh")
	err = e.writeFile(deleteJumpboxScript, deleteEnvCmd, 0750)
	if err != nil {
		return err
	}

	return nil
}

func (e Executor) getDirectorSetupFiles(input InterpolateInput) []setupFile {
	files := e.getSetupFiles(boshDeploymentRepo, input.DeploymentDir)

	statePath := filepath.Join(input.StateDir, "bbl-ops-files", input.IAAS)
	assetPath := filepath.Join(boshDeploymentRepo, input.IAAS)

	if input.IAAS == "gcp" {
		files = append(files, setupFile{
			source:   filepath.Join(assetPath, "bosh-director-ephemeral-ip-ops.yml"),
			dest:     filepath.Join(statePath, "bosh-director-ephemeral-ip-ops.yml"),
			contents: []byte(GCPBoshDirectorEphemeralIPOps),
		})
	}
	if input.IAAS == "aws" {
		files = append(files, setupFile{
			source:   filepath.Join(assetPath, "bosh-director-ephemeral-ip-ops.yml"),
			dest:     filepath.Join(statePath, "bosh-director-ephemeral-ip-ops.yml"),
			contents: []byte(AWSBoshDirectorEphemeralIPOps),
		})
		files = append(files, setupFile{
			source:   filepath.Join(assetPath, "bosh-director-encrypt-disk-ops.yml"),
			dest:     filepath.Join(statePath, "bosh-director-encrypt-disk-ops.yml"),
			contents: []byte(AWSEncryptDiskOps),
		})
	}

	return files
}

func (e Executor) getDirectorOpsFiles(input InterpolateInput) []string {
	files := []string{
		filepath.Join(input.DeploymentDir, input.IAAS, "cpi.yml"),
		filepath.Join(input.DeploymentDir, "jumpbox-user.yml"),
		filepath.Join(input.DeploymentDir, "uaa.yml"),
		filepath.Join(input.DeploymentDir, "credhub.yml"),
	}
	if input.IAAS == "gcp" {
		files = append(files, filepath.Join(input.StateDir, "bbl-ops-files", input.IAAS, "bosh-director-ephemeral-ip-ops.yml"))
	}
	if input.IAAS == "aws" {
		files = append(files, filepath.Join(input.StateDir, "bbl-ops-files", input.IAAS, "bosh-director-ephemeral-ip-ops.yml"))
		files = append(files, filepath.Join(input.DeploymentDir, input.IAAS, "iam-instance-profile.yml"))
		files = append(files, filepath.Join(input.StateDir, "bbl-ops-files", input.IAAS, "bosh-director-encrypt-disk-ops.yml"))
	}
	return files
}

func (e Executor) PlanDirector(input InterpolateInput) error {
	setupFiles := e.getDirectorSetupFiles(input)

	for _, f := range setupFiles {
		if f.source != "" {
			os.MkdirAll(filepath.Dir(f.dest), storage.StateMode)
		}
		if err := e.writeFile(f.dest, f.contents, storage.StateMode); err != nil {
			return fmt.Errorf("Director write setup file: %s", err) //not tested
		}
	}

	sharedArgs := []string{
		"--vars-store", filepath.Join(input.VarsDir, "director-vars-store.yml"),
		"--vars-file", filepath.Join(input.VarsDir, "director-vars-file.yml"),
	}

	for _, f := range e.getDirectorOpsFiles(input) {
		sharedArgs = append(sharedArgs, "-o", f)
	}

	if input.IAAS == "vsphere" {
		sharedArgs = append(sharedArgs, "-o", filepath.Join(input.DeploymentDir, "vsphere", "resource-pool.yml"))
	}

	boshState := filepath.Join(input.VarsDir, "bosh-state.json")

	boshPath, err := e.command.GetBOSHPath()
	if err != nil {
		return fmt.Errorf("Director get BOSH path: %s", err) //not tested
	}

	boshArgs := append([]string{
		filepath.Join(input.DeploymentDir, "bosh.yml"),
		"--state", boshState,
	}, sharedArgs...)

	createEnvCmd := []byte(formatScript(boshPath, input.StateDir, "create-env", boshArgs))
	err = e.writeFile(filepath.Join(input.StateDir, "create-director.sh"), createEnvCmd, 0750)
	if err != nil {
		return err
	}

	deleteEnvCmd := []byte(formatScript(boshPath, input.StateDir, "delete-env", boshArgs))
	err = e.writeFile(filepath.Join(input.StateDir, "delete-director.sh"), deleteEnvCmd, 0750)
	if err != nil {
		return err
	}

	return nil
}

func formatScript(boshPath, stateDir, command string, args []string) string {
	script := fmt.Sprintf("#!/bin/sh\n%s %s \\\n", boshPath, command)
	for _, arg := range args {
		if arg[0] == '-' {
			script = fmt.Sprintf("%s  %s", script, arg)
		} else {
			script = fmt.Sprintf("%s  %s \\\n", script, arg)
		}
	}
	script = strings.Replace(script, stateDir, "${BBL_STATE_DIR}", -1)
	return fmt.Sprintf("%s\n", script[:len(script)-2])
}

func (e Executor) WriteDeploymentVars(createEnvInput CreateEnvInput) error {
	varsFilePath := filepath.Join(createEnvInput.VarsDir, fmt.Sprintf("%s-vars-file.yml", createEnvInput.Deployment))
	err := e.writeFile(varsFilePath, []byte(createEnvInput.DeploymentVars), storage.StateMode)
	if err != nil {
		return fmt.Errorf("Write vars file: %s", err) // not tested
	}
	return nil
}

func (e Executor) CreateEnv(createEnvInput CreateEnvInput) (string, error) {
	os.Setenv("BBL_STATE_DIR", createEnvInput.StateDir)
	createEnvScript := filepath.Join(createEnvInput.StateDir, fmt.Sprintf("create-%s-override.sh", createEnvInput.Deployment))
	_, err := os.Stat(createEnvScript)
	if err != nil {
		createEnvScript = strings.Replace(createEnvScript, "-override", "", -1)
	}

	cmd := exec.Command(createEnvScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Run bosh create-env: %s", err)
	}

	varsStoreFileName := fmt.Sprintf("%s-vars-store.yml", createEnvInput.Deployment)
	varsStoreContents, err := e.readFile(filepath.Join(createEnvInput.VarsDir, varsStoreFileName))
	if err != nil {
		return "", fmt.Errorf("Reading vars file for %s deployment: %s", createEnvInput.Deployment, err) // not tested
	}

	return string(varsStoreContents), nil
}

func (e Executor) DeleteEnv(deleteEnvInput DeleteEnvInput) error {
	os.Setenv("BBL_STATE_DIR", deleteEnvInput.StateDir)
	deleteEnvScript := filepath.Join(deleteEnvInput.StateDir, fmt.Sprintf("delete-%s-override.sh", deleteEnvInput.Deployment))
	_, err := os.Stat(deleteEnvScript)
	if err != nil {
		deleteEnvScript = strings.Replace(deleteEnvScript, "-override", "", -1)
	}

	cmd := exec.Command(deleteEnvScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Run bosh delete-env: %s", err)
	}

	return nil
}

func (e Executor) Version() (string, error) {
	args := []string{"-v"}
	buffer := bytes.NewBuffer([]byte{})
	err := e.command.Run(buffer, "", args)
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
