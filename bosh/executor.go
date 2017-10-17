package bosh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

type Executor struct {
	command       command
	readFile      func(string) ([]byte, error)
	unmarshalJSON func([]byte, interface{}) error
	marshalJSON   func(interface{}) ([]byte, error)
	writeFile     func(string, []byte, os.FileMode) error
}

type InterpolateInput struct {
	DeploymentDir  string
	VarsDir        string
	IAAS           string
	DeploymentVars string
	BOSHState      map[string]interface{}
	Variables      string
	OpsFile        string
}

type InterpolateOutput struct {
	Args []string
}

type CreateEnvInput struct {
	Args       []string
	Directory  string
	Deployment string
}

type DeleteEnvInput struct {
	Args       []string
	Deployment string
	Directory  string
}

type command interface {
	Run(stdout io.Writer, workingDirectory string, args []string) error
}

const VERSION_DEV_BUILD = "[DEV BUILD]"

func NewExecutor(cmd command, readFile func(string) ([]byte, error),
	unmarshalJSON func([]byte, interface{}) error,
	marshalJSON func(interface{}) ([]byte, error), writeFile func(string, []byte, os.FileMode) error) Executor {
	return Executor{
		command:       cmd,
		readFile:      readFile,
		unmarshalJSON: unmarshalJSON,
		marshalJSON:   marshalJSON,
		writeFile:     writeFile,
	}
}

func (e Executor) JumpboxCreateEnvArgs(input InterpolateInput) (InterpolateOutput, error) {
	type setupFile struct {
		path     string
		contents []byte
	}

	setupFiles := map[string]setupFile{
		"manifest": setupFile{
			path:     filepath.Join(input.DeploymentDir, "jumpbox.yml"),
			contents: MustAsset("vendor/github.com/cppforlife/jumpbox-deployment/jumpbox.yml"),
		},
		"vars-file": setupFile{
			path:     filepath.Join(input.VarsDir, "jumpbox-deployment-vars.yml"),
			contents: []byte(input.DeploymentVars),
		},
		"cpi": setupFile{
			path:     filepath.Join(input.DeploymentDir, "cpi.yml"),
			contents: MustAsset(filepath.Join("vendor/github.com/cppforlife/jumpbox-deployment", input.IAAS, "cpi.yml")),
		},
		"vars-store": setupFile{
			path:     filepath.Join(input.VarsDir, "jumpbox-variables.yml"),
			contents: []byte(input.Variables),
		},
	}

	for _, f := range setupFiles {
		err := e.writeFile(f.path, f.contents, os.ModePerm)
		if err != nil {
			return InterpolateOutput{}, fmt.Errorf("Jumpbox write setup file: %s", err) //not tested
		}
	}

	sharedArgs := []string{
		"--vars-store", setupFiles["vars-store"].path,
		"--vars-file", setupFiles["vars-file"].path,
		"-o", setupFiles["cpi"].path,
	}

	jumpboxState := filepath.Join(input.VarsDir, "jumpbox-state.json")
	if input.BOSHState != nil {
		stateJSON, err := e.marshalJSON(input.BOSHState)
		if err != nil {
			return InterpolateOutput{}, fmt.Errorf("Jumpbox marshal state json: %s", err) //not tested
		}

		err = e.writeFile(jumpboxState, stateJSON, os.ModePerm)
		if err != nil {
			return InterpolateOutput{}, fmt.Errorf("Jumpbox write state json: %s", err) //not tested
		}
	}

	createEnvArgs := append([]string{
		"create-env", setupFiles["manifest"].path,
		"--state", jumpboxState,
	}, sharedArgs...)
	return InterpolateOutput{
		Args: createEnvArgs,
	}, nil
}

func (e Executor) DirectorCreateEnvArgs(input InterpolateInput) (InterpolateOutput, error) {
	type setupFile struct {
		path     string
		contents []byte
	}

	setupFiles := map[string]setupFile{
		"manifest": setupFile{
			path:     filepath.Join(input.DeploymentDir, "bosh.yml"),
			contents: MustAsset("vendor/github.com/cloudfoundry/bosh-deployment/bosh.yml"),
		},
		"vars-file": setupFile{
			path:     filepath.Join(input.VarsDir, "director-deployment-vars.yml"),
			contents: []byte(input.DeploymentVars),
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

	sharedArgs := []string{
		"--vars-store", setupFiles["vars-store"].path,
		"--vars-file", setupFiles["vars-file"].path,
	}

	for _, f := range opsFiles {
		sharedArgs = append(sharedArgs, "-o", f.path)
	}

	if input.OpsFile != "" {
		sharedArgs = append(sharedArgs, "-o", filepath.Join(input.VarsDir, "user-ops-file.yml"))
	}

	boshState := filepath.Join(input.VarsDir, "bosh-state.json")
	if input.BOSHState != nil {
		stateJSON, err := e.marshalJSON(input.BOSHState)
		if err != nil {
			return InterpolateOutput{}, fmt.Errorf("marshal JSON: %s", err) //not tested
		}

		err = e.writeFile(boshState, stateJSON, os.ModePerm)
		if err != nil {
			return InterpolateOutput{}, fmt.Errorf("write file: %s", err) //not tested
		}
	}

	createEnvArgs := append([]string{
		"create-env", setupFiles["manifest"].path,
		"--state", boshState,
	}, sharedArgs...)
	return InterpolateOutput{
		Args: createEnvArgs,
	}, nil
}

func (e Executor) CreateEnv(createEnvInput CreateEnvInput) (string, error) {
	err := e.command.Run(os.Stdout, createEnvInput.Directory, createEnvInput.Args)
	if err != nil {
		return "", fmt.Errorf("Create env: %s", err)
	}

	varsStoreFileName := fmt.Sprintf("%s-variables.yml", createEnvInput.Deployment)
	varsStoreContents, err := e.readFile(filepath.Join(createEnvInput.Directory, varsStoreFileName))
	if err != nil {
		return "", fmt.Errorf("Reading vars file for %s deployment: %s", createEnvInput.Deployment, err) // not tested
	}

	return string(varsStoreContents), nil
}

func (e Executor) DeleteEnv(deleteEnvInput DeleteEnvInput) error {
	deleteEnvArgs := []string{}

	for _, arg := range deleteEnvInput.Args {
		if arg == "create-env" {
			arg = "delete-env"
		}
		deleteEnvArgs = append(deleteEnvArgs, arg)
	}

	err := e.command.Run(os.Stdout, deleteEnvInput.Directory, deleteEnvArgs)
	if err != nil {
		return fmt.Errorf("Delete env: %s", err)
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
