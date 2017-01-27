package bosh

import (
	"fmt"
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
}

type ExecutorOutput struct {
	Variables map[string]interface{}
	BOSHState map[string]interface{}
}

type ExecutorInput struct {
	IAAS                  string
	Command               string
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
	Variables             string
	BOSHState             map[string]interface{}
}

type command interface {
	Run(workingDirectory string, args []string) error
}

func NewExecutor(cmd command, tempDir func(string, string) (string, error), readFile func(string) ([]byte, error),
	unmarshalYAML func([]byte, interface{}) error, unmarshalJSON func([]byte, interface{}) error,
	marshalJSON func(interface{}) ([]byte, error), writeFile func(string, []byte, os.FileMode) error) Executor {
	return Executor{
		command:       cmd,
		tempDir:       tempDir,
		readFile:      readFile,
		unmarshalYAML: unmarshalYAML,
		unmarshalJSON: unmarshalJSON,
		marshalJSON:   marshalJSON,
		writeFile:     writeFile,
	}
}

func (e Executor) Execute(executorInput ExecutorInput) (ExecutorOutput, error) {
	tempDir, err := e.tempDir("", "")
	if err != nil {
		return ExecutorOutput{}, err
	}

	statePath := fmt.Sprintf("%s/state.json", tempDir)
	variablesPath := fmt.Sprintf("%s/variables.yml", tempDir)
	privateKeyPath := filepath.Join(tempDir, "private_key")
	boshManifestPath := filepath.Join(tempDir, "bosh.yml")
	cpiOpsFilePath := filepath.Join(tempDir, "cpi.yml")
	externalIPNotRecommendedOpsFilePath := filepath.Join(tempDir, "external-ip-not-recommended.yml")

	if executorInput.BOSHState != nil {
		boshStateContents, err := e.marshalJSON(executorInput.BOSHState)
		if err != nil {
			return ExecutorOutput{}, err
		}
		err = e.writeFile(statePath, boshStateContents, os.ModePerm)
		if err != nil {
			return ExecutorOutput{}, err
		}
	}

	if executorInput.Variables != "" {
		err = e.writeFile(variablesPath, []byte(executorInput.Variables), os.ModePerm)
		if err != nil {
			return ExecutorOutput{}, err
		}
	}

	boshManifestContents, err := Asset("vendor/github.com/cloudfoundry/bosh-deployment/bosh.yml")
	if err != nil {
		//not tested
		return ExecutorOutput{}, err
	}
	err = e.writeFile(boshManifestPath, boshManifestContents, os.ModePerm)
	if err != nil {
		return ExecutorOutput{}, err
	}

	cpiOpsFileContents, err := Asset(fmt.Sprintf("vendor/github.com/cloudfoundry/bosh-deployment/%s/cpi.yml", executorInput.IAAS))
	if err != nil {
		//not tested
		return ExecutorOutput{}, err
	}
	err = e.writeFile(cpiOpsFilePath, cpiOpsFileContents, os.ModePerm)
	if err != nil {
		return ExecutorOutput{}, err
	}

	var externalIPNotRecommendedOpsFileContents []byte
	switch executorInput.IAAS {
	case "gcp":
		externalIPNotRecommendedOpsFileContents, err = Asset("vendor/github.com/cloudfoundry/bosh-deployment/external-ip-not-recommended.yml")
		if err != nil {
			//not tested
			return ExecutorOutput{}, err
		}
	case "aws":
		externalIPNotRecommendedOpsFileContents, err = Asset("vendor/github.com/cloudfoundry/bosh-deployment/external-ip-with-registry-not-recommended.yml")
		if err != nil {
			//not tested
			return ExecutorOutput{}, err
		}
	}
	err = e.writeFile(externalIPNotRecommendedOpsFilePath, externalIPNotRecommendedOpsFileContents, os.ModePerm)
	if err != nil {
		return ExecutorOutput{}, err
	}

	err = e.writeFile(privateKeyPath, []byte(executorInput.PrivateKey), os.ModePerm)
	if err != nil {
		return ExecutorOutput{}, err
	}

	args := []string{
		executorInput.Command, boshManifestPath,
		"--state", statePath,
		"-o", cpiOpsFilePath,
		"-o", externalIPNotRecommendedOpsFilePath,
		"--vars-store", variablesPath,
		"-v", "internal_cidr=10.0.0.0/24",
		"-v", "internal_gw=10.0.0.1",
		"-v", "internal_ip=10.0.0.6",
		"-v", fmt.Sprintf("external_ip=%s", executorInput.ExternalIP),
		"-v", fmt.Sprintf("director_name=%s", executorInput.DirectorName),
		"--var-file", fmt.Sprintf("private_key=%s", privateKeyPath),
	}

	switch executorInput.IAAS {
	case "gcp":
		gcpCredentialsJSONPath := filepath.Join(tempDir, "gcp_credentials.json")
		err = e.writeFile(gcpCredentialsJSONPath, []byte(executorInput.CredentialsJSON), os.ModePerm)
		if err != nil {
			return ExecutorOutput{}, err
		}

		tags := executorInput.Tags[0]
		for _, tag := range executorInput.Tags[1:] {
			tags = fmt.Sprintf("%s,%s", tags, tag)
		}

		args = append(args,
			"-v", fmt.Sprintf("zone=%s", executorInput.Zone),
			"-v", fmt.Sprintf("network=%s", executorInput.Network),
			"-v", fmt.Sprintf("subnetwork=%s", executorInput.Subnetwork),
			"-v", fmt.Sprintf("tags=[%s]", tags),
			"-v", fmt.Sprintf("project_id=%s", executorInput.ProjectID),
			"--var-file", fmt.Sprintf("gcp_credentials_json=%s", gcpCredentialsJSONPath),
		)
	case "aws":
		args = append(args,
			"-v", fmt.Sprintf("access_key_id=%s", executorInput.AccessKeyID),
			"-v", fmt.Sprintf("secret_access_key=%s", executorInput.SecretAccessKey),
			"-v", fmt.Sprintf("region=%s", executorInput.Region),
			"-v", fmt.Sprintf("az=%s", executorInput.AZ),
			"-v", fmt.Sprintf("default_key_name=%s", executorInput.DefaultKeyName),
			"-v", fmt.Sprintf("default_security_groups=%s", executorInput.DefaultSecurityGroups),
			"-v", fmt.Sprintf("subnet_id=%s", executorInput.SubnetID),
		)
	}

	err = e.command.Run(tempDir, args)
	if err != nil {
		return ExecutorOutput{}, err
	}

	variables := map[string]interface{}{}
	state := map[string]interface{}{}
	if executorInput.Command == "create-env" {
		variablesContents, err := e.readFile(variablesPath)
		if err != nil {
			return ExecutorOutput{}, err
		}
		err = e.unmarshalYAML(variablesContents, &variables)
		if err != nil {
			return ExecutorOutput{}, err
		}

		stateContents, err := e.readFile(statePath)
		if err != nil {
			return ExecutorOutput{}, err
		}
		err = e.unmarshalJSON(stateContents, &state)
		if err != nil {
			return ExecutorOutput{}, err
		}
	}

	return ExecutorOutput{
		BOSHState: state,
		Variables: variables,
	}, nil
}
