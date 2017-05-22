package commands

import yaml "gopkg.in/yaml.v2"

func SetMarshal(f func(interface{}) ([]byte, error)) {
	marshal = f
}

func ResetMarshal() {
	marshal = yaml.Marshal
}

func SetUnmarshal(f func([]byte, interface{}) error) {
	unmarshal = f
}

func ResetUnmarshal() {
	unmarshal = yaml.Unmarshal
}
