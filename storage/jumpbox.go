package storage

import "reflect"

type JumpboxDeployment struct {
	Variables string                 `json:"variables"`
	Manifest  string                 `json:"manifest"`
	State     map[string]interface{} `json:"state"`
}

func (b JumpboxDeployment) IsEmpty() bool {
	return reflect.DeepEqual(b, JumpboxDeployment{})
}
