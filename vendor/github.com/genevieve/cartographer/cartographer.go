package cartographer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type Cartographer struct{}

func NewCartographer() Cartographer {
	return Cartographer{}
}

// Ymlize reads the outputs from terraform.tfstate
// specified at path and returns yml.
func (c Cartographer) Ymlize(path string) (string, error) {
	outputs, err := c.outputs(path)
	if err != nil {
		return "", err
	}

	yml := map[string]interface{}{}
	for name, output := range outputs {
		yml[name] = output.Value
	}

	output, err := yaml.Marshal(yml)
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (c Cartographer) GetMap(path string) (map[string]interface{}, error) {
	outputs, err := c.outputs(path)
	if err != nil {
		return nil, err
	}

	yml := map[string]interface{}{}
	for name, output := range outputs {
		yml[name] = output.Value
	}

	return yml, nil
}

// Ymlize reads the terraform.tfstate specified at the path.
// It returns yml for the outputs that contain that prefix
// or outputs that have no prefix of the form `prefix__name`.
func (c Cartographer) YmlizeWithPrefix(path, prefix string) (string, error) {
	outputs, err := c.outputs(path)
	if err != nil {
		return "", err
	}

	yml := map[string]interface{}{}

	for name, output := range outputs {
		if strings.Contains(name, "__") {
			if strings.Contains(name, prefix) {
				filter := fmt.Sprintf("%s__", prefix)
				name = strings.TrimPrefix(name, filter)
				yml[name] = output.Value
			}
		} else {
			yml[name] = output.Value
		}
	}

	output, err := yaml.Marshal(yml)
	if err != nil {
		return "", fmt.Errorf("Yaml marshal: %s", err)
	}

	return string(output), nil
}

func (c Cartographer) outputs(path string) (map[string]output, error) {
	cont, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Read terraform.tfstate: %s", err)
	}

	var state tfstate
	err = json.Unmarshal(cont, &state)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal terraform.tfstate: %s", err)
	}

	if len(state.Modules) == 0 {
		return nil, errors.New("No modules found.")
	}

	return state.Modules[0].Outputs, nil
}
