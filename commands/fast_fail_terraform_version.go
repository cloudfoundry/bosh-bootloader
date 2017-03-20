package commands

import (
	"errors"

	"github.com/coreos/go-semver/semver"
)

func fastFailTerraformVersion(terraformExecutor terraformExecutor) error {
	version, err := terraformExecutor.Version()
	if err != nil {
		return err
	}

	currentVersion, err := semver.NewVersion(version)
	if err != nil {
		return err
	}

	// This shouldn't fail, so there is no test for capturing the error.
	minimumVersion, err := semver.NewVersion("0.8.5")
	if err != nil {
		return err
	}

	if currentVersion.LessThan(*minimumVersion) {
		return errors.New("Terraform version must be at least v0.8.5")
	}

	return nil
}
