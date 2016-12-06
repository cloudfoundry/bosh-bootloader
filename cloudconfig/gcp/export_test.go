package gcp

import yaml "gopkg.in/yaml.v2"

func SetUnmarshal(f func([]byte, interface{}) error) {
	unmarshal = f
}

func ResetUnmarshal() {
	unmarshal = yaml.Unmarshal
}
