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

type Executor struct {
	command       command
	tempDir       func(string, string) (string, error)
	readFile      func(string) ([]byte, error)
	unmarshalJSON func([]byte, interface{}) error
	marshalJSON   func(interface{}) ([]byte, error)
	writeFile     func(string, []byte, os.FileMode) error
}

type InterpolateInput struct {
	DeploymentDir          string
	VarsDir                string
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
	Deployment string
	Directory  string
	Manifest   string
	Variables  string
	State      map[string]interface{}
}

type CreateEnvOutput struct {
	State map[string]interface{}
}

type DeleteEnvInput struct {
	Deployment string
	Directory  string
	Manifest   string
	Variables  string
	State      map[string]interface{}
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

func (e Executor) JumpboxInterpolate(input InterpolateInput) (JumpboxInterpolateOutput, error) {
	// azure cpi is not yet in jumpbox-deployment repo
	var cpiContents []byte
	if input.IAAS == "azure" {
		cpiContents = []byte(AzureJumpboxCpi)
	} else {
		cpiContents = MustAsset(filepath.Join("vendor/github.com/cppforlife/jumpbox-deployment", input.IAAS, "cpi.yml"))
	}

	type setupFile struct {
		path     string
		contents []byte
	}

	var setupFiles = map[string]setupFile{
		"manifest": setupFile{
			path:     filepath.Join(input.DeploymentDir, "jumpbox.yml"),
			contents: MustAsset("vendor/github.com/cppforlife/jumpbox-deployment/jumpbox.yml"),
		},
		"vars-file": setupFile{
			path:     filepath.Join(input.VarsDir, "jumpbox-deployment-vars.yml"),
			contents: []byte(input.JumpboxDeploymentVars),
		},
		"cpi": setupFile{
			path:     filepath.Join(input.DeploymentDir, "cpi.yml"),
			contents: cpiContents,
		},
		"vars-store": setupFile{
			path:     filepath.Join(input.VarsDir, "jumpbox-variables.yml"),
			contents: []byte(input.Variables),
		},
	}

	for _, f := range setupFiles {
		err := e.writeFile(f.path, f.contents, os.ModePerm)
		if err != nil {
			return JumpboxInterpolateOutput{}, fmt.Errorf("write file: %s", err) //not tested
		}
	}

	args := []string{
		"interpolate", setupFiles["manifest"].path,
		"--var-errs",
		"--vars-store", setupFiles["vars-store"].path,
		"--vars-file", setupFiles["vars-file"].path,
		"-o", setupFiles["cpi"].path,
	}

	buffer := bytes.NewBuffer([]byte{})
	err := e.command.Run(buffer, input.VarsDir, args)
	if err != nil {
		return JumpboxInterpolateOutput{}, fmt.Errorf("Jumpbox interpolate: %s: %s", err, buffer)
	}

	varsStore, err := e.readFile(setupFiles["vars-store"].path)
	if err != nil {
		return JumpboxInterpolateOutput{}, fmt.Errorf("Jumpbox read file: %s", err)
	}

	return JumpboxInterpolateOutput{
		Variables: string(varsStore),
		Manifest:  buffer.String(),
	}, nil
}

func (e Executor) DirectorInterpolate(input InterpolateInput) (InterpolateOutput, error) {
	type setupFile struct {
		path     string
		contents []byte
	}

	var setupFiles = map[string]setupFile{
		"manifest": setupFile{
			path:     filepath.Join(input.DeploymentDir, "bosh.yml"),
			contents: MustAsset("vendor/github.com/cloudfoundry/bosh-deployment/bosh.yml"),
		},
		"vars-file": setupFile{
			path:     filepath.Join(input.VarsDir, "director-deployment-vars.yml"),
			contents: []byte(input.DirectorDeploymentVars),
		},
		"vars-store": setupFile{
			path:     filepath.Join(input.VarsDir, "director-variables.yml"),
			contents: []byte(input.Variables),
		},
		"user-ops": setupFile{
			path:     filepath.Join(input.VarsDir, "user-ops-file.yml"),
			contents: []byte(input.OpsFile),
		},
	}

	opsFiles := []setupFile{
		setupFile{
			path:     filepath.Join(input.DeploymentDir, "cpi.yml"),
			contents: MustAsset(filepath.Join("vendor/github.com/cloudfoundry/bosh-deployment", input.IAAS, "cpi.yml")),
		},
		setupFile{
			path:     filepath.Join(input.DeploymentDir, "jumpbox-user.yml"),
			contents: MustAsset("vendor/github.com/cloudfoundry/bosh-deployment/jumpbox-user.yml"),
		},
		setupFile{
			path:     filepath.Join(input.DeploymentDir, "uaa.yml"),
			contents: MustAsset("vendor/github.com/cloudfoundry/bosh-deployment/uaa.yml"),
		},
		setupFile{
			path:     filepath.Join(input.DeploymentDir, "credhub.yml"),
			contents: MustAsset("vendor/github.com/cloudfoundry/bosh-deployment/credhub.yml"),
		},
	}

	switch input.IAAS {
	case "gcp":
		opsFiles = append(opsFiles, setupFile{
			path:     filepath.Join(input.DeploymentDir, "gcp-bosh-director-ephemeral-ip-ops.yml"),
			contents: []byte(GCPBoshDirectorEphemeralIPOps),
		})
	case "aws":
		opsFiles = append(opsFiles,
			setupFile{
				path:     filepath.Join(input.DeploymentDir, "aws-bosh-director-ephemeral-ip-ops.yml"),
				contents: []byte(AWSBoshDirectorEphemeralIPOps),
			},
			setupFile{
				path:     filepath.Join(input.DeploymentDir, "iam-instance-profile.yml"),
				contents: MustAsset("vendor/github.com/cloudfoundry/bosh-deployment/aws/iam-instance-profile.yml"),
			},
			setupFile{
				path:     filepath.Join(input.DeploymentDir, "aws-bosh-director-encrypt-disk-ops.yml"),
				contents: []byte(AWSEncryptDiskOps),
			})
	}

	for _, f := range setupFiles {
		err := e.writeFile(f.path, f.contents, os.ModePerm)
		if err != nil {
			return InterpolateOutput{}, fmt.Errorf("write file: %s", err) //not tested
		}
	}

	for _, f := range opsFiles {
		err := e.writeFile(f.path, f.contents, os.ModePerm)
		if err != nil {
			return InterpolateOutput{}, fmt.Errorf("write file: %s", err) //not tested
		}
	}

	var args = []string{
		"interpolate", setupFiles["manifest"].path,
		"--var-errs",
		"--var-errs-unused",
		"--vars-store", setupFiles["vars-store"].path,
		"--vars-file", setupFiles["vars-file"].path,
	}

	for _, f := range opsFiles {
		args = append(args, "-o", f.path)
	}

	buffer := bytes.NewBuffer([]byte{})
	err := e.command.Run(buffer, input.VarsDir, args)
	if err != nil {
		return InterpolateOutput{}, err
	}

	if input.OpsFile != "" {
		err = e.writeFile(setupFiles["manifest"].path, buffer.Bytes(), os.ModePerm)
		if err != nil {
			//not tested
			return InterpolateOutput{}, err
		}

		args = []string{
			"interpolate", setupFiles["manifest"].path,
			"--var-errs",
			"--vars-store", setupFiles["vars-store"].path,
			"--vars-file", setupFiles["vars-file"].path,
			"-o", filepath.Join(input.VarsDir, "user-ops-file.yml"),
		}

		buffer = bytes.NewBuffer([]byte{})
		err = e.command.Run(buffer, input.VarsDir, args)
		if err != nil {
			return InterpolateOutput{}, err
		}
	}

	varsStore, err := e.readFile(setupFiles["vars-store"].path)
	if err != nil {
		return InterpolateOutput{}, err
	}

	return InterpolateOutput{
		Variables: string(varsStore),
		Manifest:  buffer.String(),
	}, nil
}

func (e Executor) CreateEnv(createEnvInput CreateEnvInput) (CreateEnvOutput, error) {
	err := e.writePreviousFiles(createEnvInput.State, createEnvInput.Variables, createEnvInput.Manifest, createEnvInput.Directory, createEnvInput.Deployment)
	if err != nil {
		return CreateEnvOutput{}, err
	}

	statePath := filepath.Join(createEnvInput.Directory, fmt.Sprintf("%s-state.json", createEnvInput.Deployment))
	variablesPath := filepath.Join(createEnvInput.Directory, fmt.Sprintf("%s-variables.yml", createEnvInput.Deployment))
	manifestPath := filepath.Join(createEnvInput.Directory, fmt.Sprintf("%s-manifest.yml", createEnvInput.Deployment))

	args := []string{
		"create-env", manifestPath,
		"--vars-store", variablesPath,
		"--state", statePath,
	}

	err = e.command.Run(os.Stdout, createEnvInput.Directory, args)
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
	err := e.writePreviousFiles(deleteEnvInput.State, deleteEnvInput.Variables, deleteEnvInput.Manifest, deleteEnvInput.Directory, deleteEnvInput.Deployment)
	if err != nil {
		return err
	}

	statePath := filepath.Join(deleteEnvInput.Directory, fmt.Sprintf("%s-state.json", deleteEnvInput.Deployment))
	variablesPath := filepath.Join(deleteEnvInput.Directory, fmt.Sprintf("%s-variables.yml", deleteEnvInput.Deployment))
	manifestPath := filepath.Join(deleteEnvInput.Directory, fmt.Sprintf("%s-manifest.yml", deleteEnvInput.Deployment))

	args := []string{
		"delete-env", manifestPath,
		"--vars-store", variablesPath,
		"--state", statePath,
	}

	err = e.command.Run(os.Stdout, deleteEnvInput.Directory, args)
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

func (e Executor) writePreviousFiles(state map[string]interface{}, variables, manifest, directory, deployment string) error {
	statePath := filepath.Join(directory, fmt.Sprintf("%s-state.json", deployment))
	variablesPath := filepath.Join(directory, fmt.Sprintf("%s-variables.yml", deployment))
	manifestPath := filepath.Join(directory, fmt.Sprintf("%s-manifest.yml", deployment))

	if state != nil {
		stateContents, err := e.marshalJSON(state)
		if err != nil {
			return err
		}
		err = e.writeFile(statePath, stateContents, os.ModePerm)
		if err != nil {
			return err
		}
	}

	err := e.writeFile(variablesPath, []byte(variables), os.ModePerm)
	if err != nil {
		// not tested
		return err
	}

	err = e.writeFile(manifestPath, []byte(manifest), os.ModePerm)
	if err != nil {
		// not tested
		return err
	}

	return nil
}
