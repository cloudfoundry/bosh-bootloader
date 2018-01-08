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
		return awsIaasLbHelper{
			aws: NewAWS(configuration),
		}

	case "azure":
		return azureIaasLbHelper{
			azure: NewAzure(configuration),
		}

	case "gcp":
		return gcpIaasLbHelper{
			gcp: NewGCP(configuration),
		}
	default:
		panic(fmt.Sprintf("%s is not a supported iaas", iaas))
	}
}
