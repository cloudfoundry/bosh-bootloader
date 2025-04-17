package azure

import "gopkg.in/yaml.v2"

func SetMarshal(f func(interface{}) ([]byte, error)) {
	marshal = f
}

func ResetMarshal() {
	marshal = yaml.Marshal
}
