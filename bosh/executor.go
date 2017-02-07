package bosh

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Executor struct {
	command       command
	tempDir       func(string, string) (string, error)
	readFile      func(string) ([]byte, error)
	unmarshalYAML func([]byte, interface{}) error
	unmarshalJSON func([]byte, interface{}) error
	marshalJSON   func(interface{}) ([]byte, error)
	writeFile     func(string, []byte, os.FileMode) error
	debug         bool
}

type InterpolateInput struct {
	IAAS                  string
	DirectorName          string
	Zone                  string
	Network               string
	Subnetwork            string
	Tags                  []string
	ProjectID             string
	ExternalIP            string
	CredentialsJSON       string
	PrivateKey            string
	DefaultKeyName        string
	DefaultSecurityGroups []string
	SubnetID              string
	AZ                    string
	Region                string
	SecretAccessKey       string
	AccessKeyID           string
	BOSHState             map[string]interface{}
	Variables             string
	OpsFile               []byte
}

type InterpolateOutput struct {
	Variables map[interface{}]interface{}
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
	Run(stdout io.Writer, workingDirectory string, args []string, debug bool) error
}

func NewExecutor(cmd command, tempDir func(string, string) (string, error), readFile func(string) ([]byte, error),
	unmarshalYAML func([]byte, interface{}) error, unmarshalJSON func([]byte, interface{}) error,
	marshalJSON func(interface{}) ([]byte, error), writeFile func(string, []byte, os.FileMode) error, debug bool) Executor {
	return Executor{
		command:       cmd,
		tempDir:       tempDir,
		readFile:      readFile,
		unmarshalYAML: unmarshalYAML,
		unmarshalJSON: unmarshalJSON,
		marshalJSON:   marshalJSON,
		writeFile:     writeFile,
		debug:         debug,
	}
}

func (e Executor) Interpolate(interpolateInput InterpolateInput) (InterpolateOutput, error) {
	tempDir, err := e.tempDir("", "")
	if err != nil {
		return InterpolateOutput{}, err
	}

	userOpsFilePath := filepath.Join(tempDir, "user-ops-file.yml")
	variablesPath := filepath.Join(tempDir, "variables.yml")
	boshManifestPath := filepath.Join(tempDir, "bosh.yml")
	cpiOpsFilePath := filepath.Join(tempDir, "cpi.yml")
	externalIPNotRecommendedOpsFilePath := filepath.Join(tempDir, "external-ip-not-recommended.yml")

	if interpolateInput.Variables != "" {
		err = e.writeFile(variablesPath, []byte(interpolateInput.Variables), os.ModePerm)
		if err != nil {
			return InterpolateOutput{}, err
		}
	}

	err = e.writeFile(userOpsFilePath, interpolateInput.OpsFile, os.ModePerm)
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

	args := []string{
		"interpolate", boshManifestPath,
		"--var-errs",
		"--var-errs-unused",
		"-o", cpiOpsFilePath,
		"-o", externalIPNotRecommendedOpsFilePath,
		"-o", userOpsFilePath,
		"--vars-store", variablesPath,
		"-v", "internal_cidr=10.0.0.0/24",
		"-v", "internal_gw=10.0.0.1",
		"-v", "internal_ip=10.0.0.6",
		"-v", fmt.Sprintf("external_ip=%s", interpolateInput.ExternalIP),
		"-v", fmt.Sprintf("director_name=%s", interpolateInput.DirectorName),
	}

	switch interpolateInput.IAAS {
	case "gcp":
		gcpCredentialsJSONPath := filepath.Join(tempDir, "gcp_credentials.json")
		err = e.writeFile(gcpCredentialsJSONPath, []byte(interpolateInput.CredentialsJSON), os.ModePerm)
		if err != nil {
			return InterpolateOutput{}, err
		}

		tags := interpolateInput.Tags[0]
		for _, tag := range interpolateInput.Tags[1:] {
			tags = fmt.Sprintf("%s,%s", tags, tag)
		}

		args = append(args,
			"-v", fmt.Sprintf("zone=%s", interpolateInput.Zone),
			"-v", fmt.Sprintf("network=%s", interpolateInput.Network),
			"-v", fmt.Sprintf("subnetwork=%s", interpolateInput.Subnetwork),
			"-v", fmt.Sprintf("tags=[%s]", tags),
			"-v", fmt.Sprintf("project_id=%s", interpolateInput.ProjectID),
			"--var-file", fmt.Sprintf("gcp_credentials_json=%s", gcpCredentialsJSONPath),
		)
	case "aws":
		privateKeyPath := filepath.Join(tempDir, "private_key")
		err = e.writeFile(privateKeyPath, []byte(interpolateInput.PrivateKey), os.ModePerm)
		if err != nil {
			return InterpolateOutput{}, err
		}

		args = append(args,
			"-v", fmt.Sprintf("access_key_id=%s", interpolateInput.AccessKeyID),
			"-v", fmt.Sprintf("secret_access_key=%s", interpolateInput.SecretAccessKey),
			"-v", fmt.Sprintf("region=%s", interpolateInput.Region),
			"-v", fmt.Sprintf("az=%s", interpolateInput.AZ),
			"-v", fmt.Sprintf("default_key_name=%s", interpolateInput.DefaultKeyName),
			"-v", fmt.Sprintf("default_security_groups=%s", interpolateInput.DefaultSecurityGroups),
			"-v", fmt.Sprintf("subnet_id=%s", interpolateInput.SubnetID),
			"--var-file", fmt.Sprintf("private_key=%s", privateKeyPath),
		)
	}

	buffer := bytes.NewBuffer([]byte{})
	err = e.command.Run(buffer, tempDir, args, true)
	if err != nil {
		return InterpolateOutput{}, err
	}

	variablesContents, err := e.readFile(variablesPath)
	if err != nil {
		return InterpolateOutput{}, err
	}

	var variables map[interface{}]interface{}
	err = e.unmarshalYAML(variablesContents, &variables)
	if err != nil {
		return InterpolateOutput{}, err
	}

	return InterpolateOutput{
		Variables: variables,
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
	boshManifestPath := filepath.Join(tempDir, "manifest.yml")

	args := []string{
		"create-env", boshManifestPath,
		"--vars-store", variablesPath,
		"--state", statePath,
	}

	err = e.command.Run(os.Stdout, tempDir, args, e.debug)
	if err != nil {
		return CreateEnvOutput{}, err
	}

	stateContents, err := e.readFile(statePath)
	if err != nil {
		return CreateEnvOutput{}, err
	}

	var state map[string]interface{}
	err = e.unmarshalJSON(stateContents, &state)
	if err != nil {
		return CreateEnvOutput{}, err
	}

	return CreateEnvOutput{
		State: state,
	}, nil
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

	err = e.command.Run(os.Stdout, tempDir, args, e.debug)
	if err != nil {
		return err
	}
	return nil
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
