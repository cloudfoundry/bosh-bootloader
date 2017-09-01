package storage

import (
	"encoding/json"

	uuid "github.com/nu7hatch/gouuid"
)

func SetMarshalIndent(f func(state interface{}, prefix, indent string) ([]byte, error)) {
	marshalIndent = f
}

func ResetMarshalIndent() {
	marshalIndent = json.MarshalIndent
}

func SetUUIDNewV4(f func() (uuid *uuid.UUID, err error)) {
	uuidNewV4 = f
}

func ResetUUIDNewV4() {
	uuidNewV4 = uuid.NewV4
}
