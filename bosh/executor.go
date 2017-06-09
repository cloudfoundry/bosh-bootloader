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
	IAAS                  string
	DeploymentVars        string
	JumpboxDeploymentVars string
	BOSHState             map[string]interface{}
	Variables             string
	OpsFile               string
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
		return JumpboxInterpolateOutput{}, err
	}

	deploymentVarsPath := filepath.Join(tempDir, "jumpbox-deployment-vars.yml")
	variablesPath := filepath.Join(tempDir, "variables.yml")
	manifestPath := filepath.Join(tempDir, "jumpbox.yml")
	cpiOpsFilePath := filepath.Join(tempDir, "cpi.yml")

	if interpolateInput.Variables != "" {
		err = e.writeFile(variablesPath, []byte(interpolateInput.Variables), os.ModePerm)
		if err != nil {
			return JumpboxInterpolateOutput{}, err
		}
	}

	err = e.writeFile(deploymentVarsPath, []byte(interpolateInput.JumpboxDeploymentVars), os.ModePerm)
	if err != nil {
		return JumpboxInterpolateOutput{}, err
	}

	manifestContents, err := Asset("vendor/github.com/cppforlife/jumpbox-deployment/jumpbox.yml")
	if err != nil {
		//not tested
		return JumpboxInterpolateOutput{}, err
	}
	err = e.writeFile(manifestPath, manifestContents, os.ModePerm)
	if err != nil {
		return JumpboxInterpolateOutput{}, err
	}

	cpiOpsFileContents, err := Asset(fmt.Sprintf("vendor/github.com/cppforlife/jumpbox-deployment/%s/cpi.yml", interpolateInput.IAAS))
	if err != nil {
		//not tested
		return JumpboxInterpolateOutput{}, err
	}
	err = e.writeFile(cpiOpsFilePath, cpiOpsFileContents, os.ModePerm)
	if err != nil {
		return JumpboxInterpolateOutput{}, err
	}

	args := []string{
		"interpolate", manifestPath,
		"--var-errs",
		"-o", cpiOpsFilePath,
		"--vars-store", variablesPath,
		"--vars-file", deploymentVarsPath,
	}

	buffer := bytes.NewBuffer([]byte{})
	err = e.command.Run(buffer, tempDir, args)
	if err != nil {
		return JumpboxInterpolateOutput{}, err
	}

	varsStore, err := e.readFile(variablesPath)
	if err != nil {
		return JumpboxInterpolateOutput{}, err
	}

	return JumpboxInterpolateOutput{
		Variables: string(varsStore),
		Manifest:  buffer.String(),
	}, nil
}

func (e Executor) Interpolate(interpolateInput InterpolateInput) (InterpolateOutput, error) {
	tempDir, err := e.tempDir("", "")
	if err != nil {
		return InterpolateOutput{}, err
	}

	deploymentVarsPath := filepath.Join(tempDir, "deployment-vars.yml")
	userOpsFilePath := filepath.Join(tempDir, "user-ops-file.yml")
	variablesPath := filepath.Join(tempDir, "variables.yml")
	boshManifestPath := filepath.Join(tempDir, "bosh.yml")
	cpiOpsFilePath := filepath.Join(tempDir, "cpi.yml")

	if interpolateInput.Variables != "" {
		err = e.writeFile(variablesPath, []byte(interpolateInput.Variables), os.ModePerm)
		if err != nil {
			return InterpolateOutput{}, err
		}
	}

	err = e.writeFile(deploymentVarsPath, []byte(interpolateInput.DeploymentVars), os.ModePerm)
	if err != nil {
		return InterpolateOutput{}, err
	}

	err = e.writeFile(userOpsFilePath, []byte(interpolateInput.OpsFile), os.ModePerm)
	if err != nil {
		return InterpolateOutput{}, err
	}

	boshManifestContents, err := Asset("vendor/github.com/cloudfoundry/bosh-deployment/bosh.yml")
	if err != nil {
		//not tested
		return InterpolateOutput{}, err
	}
	err = e.writeFile(boshManifestPath, boshManifestContents, os.ModePerm)
	if err != nil {
		return InterpolateOutput{}, err
	}

	cpiOpsFileContents, err := Asset(fmt.Sprintf("vendor/github.com/cloudfoundry/bosh-deployment/%s/cpi.yml", interpolateInput.IAAS))
	if err != nil {
		//not tested
		return InterpolateOutput{}, err
	}
	err = e.writeFile(cpiOpsFilePath, cpiOpsFileContents, os.ModePerm)
	if err != nil {
		return InterpolateOutput{}, err
	}

	var args []string

	if interpolateInput.JumpboxDeploymentVars != "" {
		args = []string{
			"interpolate", boshManifestPath,
			"--var-errs",
			"--var-errs-unused",
			"-o", cpiOpsFilePath,
			"--vars-store", variablesPath,
			"--vars-file", deploymentVarsPath,
		}
	} else {
		jumpboxUserOpsFilePath := filepath.Join(tempDir, "jumpbox-user.yml")
		jumpboxUserOpsFileContents, err := Asset("vendor/github.com/cloudfoundry/bosh-deployment/jumpbox-user.yml")
		if err != nil {
			//not tested
			return InterpolateOutput{}, err
		}
		err = e.writeFile(jumpboxUserOpsFilePath, jumpboxUserOpsFileContents, os.ModePerm)
		if err != nil {
			return InterpolateOutput{}, err
		}

		externalIPNotRecommendedOpsFilePath := filepath.Join(tempDir, "external-ip-not-recommended.yml")
		var externalIPNotRecommendedOpsFileContents []byte
		switch interpolateInput.IAAS {
		case "gcp":
			externalIPNotRecommendedOpsFileContents, err = Asset("vendor/github.com/cloudfoundry/bosh-deployment/external-ip-not-recommended.yml")
			if err != nil {
				//not tested
				return InterpolateOutput{}, err
			}
		case "aws":
			externalIPNotRecommendedOpsFileContents, err = Asset("vendor/github.com/cloudfoundry/bosh-deployment/external-ip-with-registry-not-recommended.yml")
			if err != nil {
				//not tested
				return InterpolateOutput{}, err
			}
		}
		err = e.writeFile(externalIPNotRecommendedOpsFilePath, externalIPNotRecommendedOpsFileContents, os.ModePerm)
		if err != nil {
			return InterpolateOutput{}, err
		}

		args = []string{
			"interpolate", boshManifestPath,
			"--var-errs",
			"--var-errs-unused",
			"-o", cpiOpsFilePath,
			"-o", jumpboxUserOpsFilePath,
			"-o", externalIPNotRecommendedOpsFilePath,
			"--vars-store", variablesPath,
			"--vars-file", deploymentVarsPath,
		}
	}

	buffer := bytes.NewBuffer([]byte{})
	err = e.command.Run(buffer, tempDir, args)
	if err != nil {
		return InterpolateOutput{}, err
	}

	if interpolateInput.OpsFile != "" {
		err = e.writeFile(boshManifestPath, buffer.Bytes(), os.ModePerm)
		if err != nil {
			//not tested
			return InterpolateOutput{}, err
		}

		args = []string{
			"interpolate", boshManifestPath,
			"--var-errs",
			"-o", userOpsFilePath,
			"--vars-store", variablesPath,
			"--vars-file", deploymentVarsPath,
		}

		buffer = bytes.NewBuffer([]byte{})
		err = e.command.Run(buffer, tempDir, args)
		if err != nil {
			return InterpolateOutput{}, err
		}
	}

	varsStore, err := e.readFile(variablesPath)
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

	statePath := fmt.Sprintf("%s/state.json", tempDir)
	variablesPath := fmt.Sprintf("%s/variables.yml", tempDir)
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

	statePath := fmt.Sprintf("%s/state.json", tempDir)
	variablesPath := fmt.Sprintf("%s/variables.yml", tempDir)
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

	statePath := fmt.Sprintf("%s/state.json", tempDir)
	variablesPath := fmt.Sprintf("%s/variables.yml", tempDir)
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
