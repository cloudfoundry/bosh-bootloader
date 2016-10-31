package application

import (
	"fmt"
	"os"
	"path/filepath"
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
		return fmt.Errorf("bbl-state.json not found in %q, ensure you're running this command in the proper state directory or create a new environment with bbl up", s.stateDir)
	}
	if err != nil {
		return err
	}
	return nil
}
