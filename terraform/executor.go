package terraform

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

var tempDir func(dir, prefix string) (string, error) = ioutil.TempDir
var writeFile func(file string, data []byte, perm os.FileMode) error = ioutil.WriteFile
var readFile func(filename string) ([]byte, error) = ioutil.ReadFile

type Executor struct {
	cmd terraformCmd
}

type terraformCmd interface {
	Run(stdout io.Writer, workingDirectory string, args []string) error
}

func NewExecutor(cmd terraformCmd) Executor {
	return Executor{cmd: cmd}
}

func (e Executor) Apply(credentials, envID, projectID, zone, region, template, prevTFState string) (string, error) {
	templateDir, err := tempDir("", "")
	if err != nil {
		return "", err
	}

	err = writeFile(filepath.Join(templateDir, "template.tf"), []byte(template), os.ModePerm)
	if err != nil {
		return "", err
	}

	if prevTFState != "" {
		err = writeFile(filepath.Join(templateDir, "terraform.tfstate"), []byte(prevTFState), os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	args := []string{"apply"}
	args = append(args, makeVar("project_id", projectID)...)
	args = append(args, makeVar("env_id", envID)...)
	args = append(args, makeVar("region", region)...)
	args = append(args, makeVar("zone", zone)...)
	args = append(args, makeVar("credentials", credentials)...)
	err = e.cmd.Run(os.Stdout, templateDir, args)
	if err != nil {
		return "", err
	}

	tfState, err := readFile(filepath.Join(templateDir, "terraform.tfstate"))
	if err != nil {
		return "", err
	}

	return string(tfState), nil
}

func (e Executor) Destroy(credentials, envID, projectID, zone, region, template, tfState string) error {
	tempTFDir, err := tempDir("", "")
	if err != nil {
		return err
	}

	credentialsPath := filepath.Join(tempTFDir, "credentials.json")
	err = writeFile(credentialsPath, []byte(credentials), os.ModePerm)
	if err != nil {
		return err
	}

	err = writeFile(filepath.Join(tempTFDir, "template.tf"), []byte(template), os.ModePerm)
	if err != nil {
		return err
	}

	err = writeFile(filepath.Join(tempTFDir, "terraform.tfstate"), []byte(tfState), os.ModePerm)
	if err != nil {
		return err
	}

	args := []string{"destroy"}
	args = append(args, makeVar("project_id", projectID)...)
	args = append(args, makeVar("env_id", envID)...)
	args = append(args, makeVar("region", region)...)
	args = append(args, makeVar("zone", zone)...)
	args = append(args, makeVar("credentials", credentialsPath)...)
	err = e.cmd.Run(tempTFDir, args)
	if err != nil {
		return err
	}

	return nil
}

func makeVar(name string, value string) []string {
	return []string{"-var", fmt.Sprintf("%s=%s", name, value)}
}
