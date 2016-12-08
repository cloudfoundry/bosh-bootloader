package commands

import yaml "gopkg.in/yaml.v2"

func SetMarshal(f func(in interface{}) (out []byte, err error)) {
	marshal = f
}

func ResetMarshal() {
	marshal = yaml.Marshal
}
