package commands

import (
	"errors"

	"github.com/coreos/go-semver/semver"
)

func fastFailBOSHVersion(boshManager boshManager) error {
	version, err := boshManager.Version()
	if err != nil {
		return err
	}

	currentVersion, err := semver.NewVersion(version)
	if err != nil {
		return err
	}

	// This shouldn't fail, so there is no test for capturing the error.
	minimumVersion, err := semver.NewVersion("2.0.0")
	if err != nil {
		return err
	}

	if currentVersion.LessThan(*minimumVersion) {
		return errors.New("BOSH version must be at least v2.0.0")
	}

	return nil
}
