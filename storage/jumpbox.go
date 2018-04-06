package storage

import "reflect"

type Jumpbox struct {
	URL       string                 `json:"url"`
	Variables string                 `json:"variables,omitempty"`
	Manifest  string                 `json:"manifest,omitempty"`
	State     map[string]interface{} `json:"state,omitempty"`
}

func (j Jumpbox) IsEmpty() bool {
	return reflect.DeepEqual(j, Jumpbox{})
}
