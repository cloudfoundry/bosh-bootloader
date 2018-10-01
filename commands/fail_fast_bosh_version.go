package commands

import (
	"fmt"

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

	minimumVersion, err := semver.NewVersion("2.0.48")
	if err != nil {
		return err
	}

	if currentVersion.LessThan(*minimumVersion) {
		path := boshManager.Path()
		return fmt.Errorf("%s: bosh-cli version must be at least v2.0.48, but found v%s", path, currentVersion)
	}

	return nil
}
