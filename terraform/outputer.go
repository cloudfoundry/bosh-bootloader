package terraform

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

type Outputer struct {
	cmd terraformCmd
}

func NewOutputer(cmd terraformCmd) Outputer {
	return Outputer{cmd: cmd}
}

func (o Outputer) Get(tfState, outputName string) (string, error) {
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
	err = o.cmd.Run(buffer, templateDir, args)
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(buffer.String(), "\n"), nil
}
