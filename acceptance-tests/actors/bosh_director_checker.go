package actors

import acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"

type BOSHDirectorChecker interface {
	NetworkHasBOSHDirector(string) bool
}

func NewBOSHDirectorChecker(config acceptance.Config) BOSHDirectorChecker {
	switch config.IAAS {
	case "aws":
		return NewAWS(config)
	case "azure":
		return NewAzure(config)
	case "gcp":
		return NewGCP(config)
	}
	return nil
}
