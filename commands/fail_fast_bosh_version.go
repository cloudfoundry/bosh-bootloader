package commands

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/coreos/go-semver/semver"
)

func fastFailBOSHVersion(boshManager boshManager) error {
	version, err := boshManager.Version()
	switch err.(type) {
	case bosh.BOSHVersionError:
		return nil
	case error:
		return err
	}

	currentVersion, err := semver.NewVersion(version)
	if err != nil {
		return err
	}

	// This shouldn't fail, so there is no test for capturing the error.
	minimumVersion, err := semver.NewVersion("2.0.48")
	if err != nil {
		return err
	}

	if currentVersion.LessThan(*minimumVersion) {
		return errors.New("BOSH version must be at least v2.0.48")
	}

	return nil
}
