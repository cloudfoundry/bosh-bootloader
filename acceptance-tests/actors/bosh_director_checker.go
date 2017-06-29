package actors

import acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"

type BOSHDirectorChecker interface {
	NetworkHasBOSHDirector(string) bool
}

func NewBOSHDirectorChecker(config acceptance.Config) BOSHDirectorChecker {
	switch GetIAAS(config) {
	case GCPIAAS:
		return NewGCP(config)
	case AWSIAAS:
		return NewAWS(config)
	}
	return nil
}
