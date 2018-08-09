package application

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/commands"
)

type StateValidator struct {
	stateDir string
}

func NewStateValidator(stateDir string) StateValidator {
	return StateValidator{stateDir: stateDir}
}

func (s StateValidator) Validate() error {
	_, err := os.Stat(filepath.Join(s.stateDir, "bbl-state.json"))
	if os.IsNotExist(err) {
		return commands.NewNoBBLStateError(s.stateDir)
	}
	if err != nil {
		return err
	}
	return nil
}
