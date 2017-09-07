package azure

import yaml "gopkg.in/yaml.v2"

func SetMarshal(f func(interface{}) ([]byte, error)) {
	marshal = f
}

func ResetMarshal() {
	marshal = yaml.Marshal
}
