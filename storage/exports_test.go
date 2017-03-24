package storage

import "encoding/json"

func SetMarshalIndent(f func(state interface{}, prefix, indent string) ([]byte, error)) {
	marshalIndent = f
}

func ResetMarshalIndent() {
	marshalIndent = json.MarshalIndent
}
