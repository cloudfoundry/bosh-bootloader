package storage

import (
	"fmt"
	"reflect"
	"strings"
)

type Jumpbox struct {
	URL       string                 `json:"url"`
	Variables string                 `json:"variables,omitempty"`
	Manifest  string                 `json:"manifest,omitempty"`
	State     map[string]interface{} `json:"state,omitempty"`
}

func (j Jumpbox) IsEmpty() bool {
	return reflect.DeepEqual(j, Jumpbox{})
}

func (j Jumpbox) GetURLWithJumpboxUser() string {
	if strings.Contains(j.URL, "@") {
		return j.URL
	}
	return fmt.Sprintf("jumpbox@%s", j.URL)
}
