package actors

import (
	"fmt"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
)

type IAASLBHelper interface {
	GetLBArgs() []string
	ConfirmLBsExist(string)
	ConfirmNoLBsExist(string)
	VerifyBblLBOutput(string)
	VerifyCloudConfigExtensions([]string)
	ConfirmNoStemcellsExist([]string)
}

func NewIAASLBHelper(iaas string, configuration acceptance.Config) IAASLBHelper {
	switch iaas {
	case "aws":
		return NewAWSLBHelper(configuration)
	case "azure":
		return NewAzureLBHelper(configuration)
	case "gcp":
		return NewGCPLBHelper(configuration)
	case "vsphere":
		return NewVSphereLBHelper()
	case "openstack":
		return NewOpenStackLBHelper()
	case "cloudstack":
		return NewCloudStackLBHelper()
	default:
		panic(fmt.Sprintf("%s is not a supported iaas", iaas))
	}
}
