package config

import (
	"fmt"
	"os/exec"
)

func GetBOSHPath() (string, error) {
	var boshPath = "bosh"

	path, err := exec.LookPath("bosh2")
	if err != nil {
		if err.(*exec.Error).Err != exec.ErrNotFound {
			return "", fmt.Errorf("failed when searching for BOSH: %s", err) // not tested
		}
	}

	if path != "" {
		boshPath = path
	}

	return boshPath, nil
}
