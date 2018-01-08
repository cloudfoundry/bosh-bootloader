package actors

import (
	"fmt"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
)

type IAASLBHelper interface {
	GetLBArgs() []string
	ConfirmLBsExist(string)
	ConfirmNoLBsExist(string)
}

func NewIAASLBHelper(iaas string, configuration acceptance.Config) IAASLBHelper {
	switch iaas {
	case "aws":
		return NewAWSLBHelper(configuration)
	case "azure":
		return NewAzureLBHelper(configuration)
	case "gcp":
		return NewGCPLBHelper(configuration)
	default:
		panic(fmt.Sprintf("%s is not a supported iaas", iaas))
	}
}
