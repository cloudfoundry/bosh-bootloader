package terraform

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

var tempDir func(dir, prefix string) (string, error) = ioutil.TempDir
var writeFile func(file string, data []byte, perm os.FileMode) error = ioutil.WriteFile

type Applier struct {
	cmd terraformCmd
}

type terraformCmd interface {
	Run(args []string) error
}

func NewApplier(cmd terraformCmd) Applier {
	return Applier{cmd: cmd}
}

func (applier Applier) Apply(credentials, envID, projectID, zone, region, template string) (string, error) {
	templateDir, err := tempDir("", "")
	if err != nil {
		return "", err
	}

	err = writeFile(filepath.Join(templateDir, "template.tf"), []byte(template), os.ModePerm)
	if err != nil {
		return "", err
	}

	args := []string{"apply"}
	args = append(args, makeVar("project_id", projectID)...)
	args = append(args, makeVar("env_id", envID)...)
	args = append(args, makeVar("region", region)...)
	args = append(args, makeVar("zone", zone)...)
	args = append(args, makeVar("credentials", credentials)...)
	args = append(args, templateDir)
	err = applier.cmd.Run(args)
	if err != nil {
		return "", err
	}

	return "", nil
}

func makeVar(name string, value string) []string {
	return []string{"-var", fmt.Sprintf("%s=%s", name, value)}
}
