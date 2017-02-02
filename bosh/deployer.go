package bosh

import (
	"fmt"
	"os"
	"path/filepath"
)

type Deployer struct {
	command       command
	tempDir       func(string, string) (string, error)
	readFile      func(string) ([]byte, error)
	unmarshalYAML func([]byte, interface{}) error
	unmarshalJSON func([]byte, interface{}) error
	marshalJSON   func(interface{}) ([]byte, error)
	writeFile     func(string, []byte, os.FileMode) error
}

type DeployOutput struct {
	Variables map[string]interface{}
	BOSHState map[string]interface{}
}

type DeployInput struct {
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
	Variables             string
	BOSHState             map[string]interface{}
}

type command interface {
	Run(workingDirectory string, args []string) error
}

func NewDeployer(cmd command, tempDir func(string, string) (string, error), readFile func(string) ([]byte, error),
	unmarshalYAML func([]byte, interface{}) error, unmarshalJSON func([]byte, interface{}) error,
	marshalJSON func(interface{}) ([]byte, error), writeFile func(string, []byte, os.FileMode) error) Deployer {
	return Deployer{
		command:       cmd,
		tempDir:       tempDir,
		readFile:      readFile,
		unmarshalYAML: unmarshalYAML,
		unmarshalJSON: unmarshalJSON,
		marshalJSON:   marshalJSON,
		writeFile:     writeFile,
	}
}

func (d Deployer) Deploy(deployInput DeployInput) (DeployOutput, error) {
	tempDir, err := d.tempDir("", "")
	if err != nil {
		return DeployOutput{}, err
	}

	statePath := fmt.Sprintf("%s/state.json", tempDir)
	variablesPath := fmt.Sprintf("%s/variables.yml", tempDir)
	privateKeyPath := filepath.Join(tempDir, "private_key")
	boshManifestPath := filepath.Join(tempDir, "bosh.yml")
	cpiOpsFilePath := filepath.Join(tempDir, "cpi.yml")
	externalIPNotRecommendedOpsFilePath := filepath.Join(tempDir, "external-ip-not-recommended.yml")

	if deployInput.BOSHState != nil {
		boshStateContents, err := d.marshalJSON(deployInput.BOSHState)
		if err != nil {
			return DeployOutput{}, err
		}
		err = d.writeFile(statePath, boshStateContents, os.ModePerm)
		if err != nil {
			return DeployOutput{}, err
		}
	}

	if deployInput.Variables != "" {
		err = d.writeFile(variablesPath, []byte(deployInput.Variables), os.ModePerm)
		if err != nil {
			return DeployOutput{}, err
		}
	}

	boshManifestContents, err := Asset("vendor/github.com/cloudfoundry/bosh-deployment/bosh.yml")
	if err != nil {
		//not tested
		return DeployOutput{}, err
	}
	err = d.writeFile(boshManifestPath, boshManifestContents, os.ModePerm)
	if err != nil {
		return DeployOutput{}, err
	}

	cpiOpsFileContents, err := Asset("vendor/github.com/cloudfoundry/bosh-deployment/gcp/cpi.yml")
	if err != nil {
		//not tested
		return DeployOutput{}, err
	}
	err = d.writeFile(cpiOpsFilePath, cpiOpsFileContents, os.ModePerm)
	if err != nil {
		return DeployOutput{}, err
	}

	var externalIPNotRecommendedOpsFileContents []byte
	switch deployInput.IAAS {
	case "gcp":
		externalIPNotRecommendedOpsFileContents, err = Asset("vendor/github.com/cloudfoundry/bosh-deployment/external-ip-not-recommended.yml")
		if err != nil {
			//not tested
			return DeployOutput{}, err
		}
	case "aws":
		externalIPNotRecommendedOpsFileContents, err = Asset("vendor/github.com/cloudfoundry/bosh-deployment/external-ip-with-registry-not-recommended.yml")
		if err != nil {
			//not tested
			return DeployOutput{}, err
		}
	}
	err = d.writeFile(externalIPNotRecommendedOpsFilePath, externalIPNotRecommendedOpsFileContents, os.ModePerm)
	if err != nil {
		return DeployOutput{}, err
	}

	err = d.writeFile(privateKeyPath, []byte(deployInput.PrivateKey), os.ModePerm)
	if err != nil {
		return DeployOutput{}, err
	}

	args := []string{
		"create-env", boshManifestPath,
		"--state", statePath,
		"-o", cpiOpsFilePath,
		"-o", externalIPNotRecommendedOpsFilePath,
		"--vars-store", variablesPath,
		"-v", "internal_cidr=10.0.0.0/24",
		"-v", "internal_gw=10.0.0.1",
		"-v", "internal_ip=10.0.0.6",
		"-v", fmt.Sprintf("external_ip=%s", deployInput.ExternalIP),
		"-v", fmt.Sprintf("director_name=%s", deployInput.DirectorName),
		"--var-file", fmt.Sprintf("private_key=%s", privateKeyPath),
	}

	switch deployInput.IAAS {
	case "gcp":
		gcpCredentialsJSONPath := filepath.Join(tempDir, "gcp_credentials.json")
		err = d.writeFile(gcpCredentialsJSONPath, []byte(deployInput.CredentialsJSON), os.ModePerm)
		if err != nil {
			return DeployOutput{}, err
		}

		tags := deployInput.Tags[0]
		for _, tag := range deployInput.Tags[1:] {
			tags = fmt.Sprintf("%s,%s", tags, tag)
		}

		args = append(args,
			"-v", fmt.Sprintf("zone=%s", deployInput.Zone),
			"-v", fmt.Sprintf("network=%s", deployInput.Network),
			"-v", fmt.Sprintf("subnetwork=%s", deployInput.Subnetwork),
			"-v", fmt.Sprintf("tags=[%s]", tags),
			"-v", fmt.Sprintf("project_id=%s", deployInput.ProjectID),
			"--var-file", fmt.Sprintf("gcp_credentials_json=%s", gcpCredentialsJSONPath),
		)
	case "aws":
		args = append(args,
			"-v", fmt.Sprintf("access_key_id=%s", deployInput.AccessKeyID),
			"-v", fmt.Sprintf("secret_access_key=%s", deployInput.SecretAccessKey),
			"-v", fmt.Sprintf("region=%s", deployInput.Region),
			"-v", fmt.Sprintf("az=%s", deployInput.AZ),
			"-v", fmt.Sprintf("default_key_name=%s", deployInput.DefaultKeyName),
			"-v", fmt.Sprintf("default_security_groups=%s", deployInput.DefaultSecurityGroups),
			"-v", fmt.Sprintf("subnet_id=%s", deployInput.SubnetID),
		)
	}

	err = d.command.Run(tempDir, args)
	if err != nil {
		return DeployOutput{}, err
	}

	variablesContents, err := d.readFile(variablesPath)
	if err != nil {
		return DeployOutput{}, err
	}
	variables := map[string]interface{}{}
	err = d.unmarshalYAML(variablesContents, &variables)
	if err != nil {
		return DeployOutput{}, err
	}

	stateContents, err := d.readFile(statePath)
	if err != nil {
		return DeployOutput{}, err
	}
	state := map[string]interface{}{}
	err = d.unmarshalJSON(stateContents, &state)
	if err != nil {
		return DeployOutput{}, err
	}

	return DeployOutput{
		BOSHState: state,
		Variables: variables,
	}, nil
}
