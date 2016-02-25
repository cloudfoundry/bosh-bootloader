package unsupported

import "fmt"

func NewIssue(message string) error {
	return fmt.Errorf("%s, please open an issue at https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/new if you require assistance.", message)
}
