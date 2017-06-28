package actors

import integration "github.com/cloudfoundry/bosh-bootloader/integration-test"

type BOSHDirectorChecker interface {
	NetworkHasBOSHDirector(string) bool
}

func NewBOSHDirectorChecker(config integration.Config) BOSHDirectorChecker {
	switch GetIAAS(config) {
	case GCPIAAS:
		return NewGCP(config)
	case AWSIAAS:
		return NewAWS(config)
	}
	return nil
}
