package terraform

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var tempDir func(dir, prefix string) (string, error) = ioutil.TempDir
var writeFile func(file string, data []byte, perm os.FileMode) error = ioutil.WriteFile
var readFile func(filename string) ([]byte, error) = ioutil.ReadFile

type Executor struct {
	cmd   terraformCmd
	debug bool
}

type terraformCmd interface {
	Run(stdout io.Writer, workingDirectory string, args []string, debug bool) error
}

func NewExecutor(cmd terraformCmd, debug bool) Executor {
	return Executor{cmd: cmd, debug: debug}
}

func (e Executor) Apply(input map[string]string, template, prevTFState string) (string, error) {
	tempDir, err := tempDir("", "")
	if err != nil {
		return "", err
	}

	err = writeFile(filepath.Join(tempDir, "template.tf"), []byte(template), os.ModePerm)
	if err != nil {
		return "", err
	}

	if prevTFState != "" {
		err = writeFile(filepath.Join(tempDir, "terraform.tfstate"), []byte(prevTFState), os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	args := []string{"apply"}
	for k, v := range input {
		args = append(args, makeVar(k, v)...)
	}
	err = e.cmd.Run(os.Stdout, tempDir, args, e.debug)
	if err != nil {
		return "", NewExecutorError(filepath.Join(tempDir, "terraform.tfstate"), err, e.debug)
	}

	tfState, err := readFile(filepath.Join(tempDir, "terraform.tfstate"))
	if err != nil {
		return "", err
	}

	return string(tfState), nil
}

func (e Executor) Destroy(input map[string]string, template, prevTFState string) (string, error) {
	tempDir, err := tempDir("", "")
	if err != nil {
		return "", err
	}

	err = writeFile(filepath.Join(tempDir, "template.tf"), []byte(template), os.ModePerm)
	if err != nil {
		return "", err
	}

	if prevTFState != "" {
		err = writeFile(filepath.Join(tempDir, "terraform.tfstate"), []byte(prevTFState), os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	args := []string{"destroy", "-force"}
	for k, v := range input {
		args = append(args, makeVar(k, v)...)
	}
	err = e.cmd.Run(os.Stdout, tempDir, args, e.debug)
	if err != nil {
		return "", NewExecutorError(filepath.Join(tempDir, "terraform.tfstate"), err, e.debug)
	}

	tfState, err := readFile(filepath.Join(tempDir, "terraform.tfstate"))
	if err != nil {
		return "", err
	}

	return string(tfState), nil
}

func (e Executor) Version() (string, error) {
	buffer := bytes.NewBuffer([]byte{})
	err := e.cmd.Run(buffer, "/tmp", []string{"version"}, true)
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

func (e Executor) Output(tfState, outputName string) (string, error) {
	templateDir, err := tempDir("", "")
	if err != nil {
		return "", err
	}

	err = writeFile(filepath.Join(templateDir, "terraform.tfstate"), []byte(tfState), os.ModePerm)
	if err != nil {
		return "", err
	}

	args := []string{"output", outputName}
	buffer := bytes.NewBuffer([]byte{})
	err = e.cmd.Run(buffer, templateDir, args, true)
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(buffer.String(), "\n"), nil

}

func makeVar(name string, value string) []string {
	return []string{"-var", fmt.Sprintf("%s=%s", name, value)}
}
