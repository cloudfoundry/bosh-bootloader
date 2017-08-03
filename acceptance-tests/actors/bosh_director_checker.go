package actors

import acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"

type BOSHDirectorChecker interface {
	NetworkHasBOSHDirector(string) bool
}

func NewBOSHDirectorChecker(config acceptance.Config) BOSHDirectorChecker {
	switch config.IAAS {
	case "gcp":
		return NewGCP(config)
	case "aws":
		return NewAWS(config)
	}
	return nil
}
