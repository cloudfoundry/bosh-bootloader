package aws

import "encoding/json"

func SetJSONMarshal(f func(interface{}) ([]byte, error)) {
	jsonMarshal = f
}

func ResetJSONMarshal() {
	jsonMarshal = json.Marshal
}
