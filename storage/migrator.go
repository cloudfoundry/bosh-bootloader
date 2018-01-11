package storage

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/cloudfoundry/bosh-bootloader/fileio"
)

type store interface {
	Set(state State) error
	GetVarsDir() (string, error)
	GetTerraformDir() (string, error)
	GetOldBblDir() string
	GetCloudConfigDir() (string, error)
}

type Migrator struct {
	store  store
	fileIO fileio.FileIO
}

func NewMigrator(store store, fileIO fileio.FileIO) Migrator {
	return Migrator{store: store, fileIO: fileIO}
}

func (m Migrator) Migrate(state State) (State, error) {
	if reflect.DeepEqual(state, State{}) {
		return state, nil
	}

	varsDir, err := m.store.GetVarsDir()
	if err != nil {
		return State{}, fmt.Errorf("migrating state: %s", err)
	}

	state, err = m.MigrateTerraformState(state, varsDir)
	if err != nil {
		return State{}, err
	}

	terraformDir, err := m.store.GetTerraformDir()
	if err != nil {
		return State{}, fmt.Errorf("migrating terraform: %s", err)
	}

	err = m.MigrateTerraformTemplate(terraformDir)
	if err != nil {
		return State{}, err
	}

	state, err = m.MigrateDirectorState(state, varsDir)
	if err != nil {
		return State{}, err
	}

	state, err = m.MigrateJumpboxState(state, varsDir)
	if err != nil {
		return State{}, err
	}

	bblDir := m.store.GetOldBblDir()
	cloudConfigDir, err := m.store.GetCloudConfigDir()
	if err != nil {
		return State{}, fmt.Errorf("getting cloud-config dir: %s", err)
	}
	err = m.MigrateCloudConfigDir(bblDir, cloudConfigDir)
	if err != nil {
		return State{}, err
	}

	err = m.MigrateTerraformVars(varsDir)
	if err != nil {
		return State{}, err
	}

	state, err = m.MigrateDirectorVars(state, varsDir)
	if err != nil {
		return State{}, err
	}

	state, err = m.MigrateJumpboxVars(state, varsDir)
	if err != nil {
		return State{}, err
	}

	err = m.store.Set(state)
	if err != nil {
		return State{}, fmt.Errorf("saving migrated state: %s", err)
	}

	return state, nil
}

func (m Migrator) MigrateTerraformState(state State, varsDir string) (State, error) {
	if state.TFState != "" {
		err := m.fileIO.WriteFile(filepath.Join(varsDir, "terraform.tfstate"), []byte(state.TFState), StateMode)
		if err != nil {
			return State{}, fmt.Errorf("migrating terraform state: %s", err)
		}
		state.TFState = ""
	}
	return state, nil
}

func (m Migrator) MigrateTerraformTemplate(terraformDir string) error {
	oldTemplatePath := filepath.Join(terraformDir, "template.tf")
	_, err := m.fileIO.Stat(oldTemplatePath)
	if err == nil {
		err = m.fileIO.Rename(oldTemplatePath, filepath.Join(terraformDir, "bbl-template.tf"))
		if err != nil {
			return fmt.Errorf("migrating terraform template: %s", err)
		}
	}
	return nil
}

func (m Migrator) migrateStateFile(state map[string]interface{}, deployment, varsDir string) error {
	if len(state) > 0 {
		stateJSON, err := json.Marshal(state)
		if err != nil {
			return fmt.Errorf("marshalling %s state: %s", deployment, err)
		}
		err = m.fileIO.WriteFile(filepath.Join(varsDir, fmt.Sprintf("%s-state.json", deployment)), stateJSON, StateMode)
		if err != nil {
			return fmt.Errorf("migrating %s state: %s", deployment, err)
		}
	}
	return nil
}

func (m Migrator) MigrateDirectorState(state State, varsDir string) (State, error) {
	err := m.migrateStateFile(state.BOSH.State, "bosh", varsDir)
	if err != nil {
		return State{}, err
	}
	state.BOSH.State = nil
	return state, nil
}

func (m Migrator) MigrateJumpboxState(state State, varsDir string) (State, error) {
	err := m.migrateStateFile(state.Jumpbox.State, "jumpbox", varsDir)
	if err != nil {
		return State{}, err
	}
	state.Jumpbox.State = nil
	return state, nil
}

func (m Migrator) MigrateCloudConfigDir(bblDir, cloudConfigDir string) error {
	if _, err := m.fileIO.Stat(bblDir); err == nil {
		oldCloudConfigDir := filepath.Join(bblDir, "cloudconfig")
		files, err := m.fileIO.ReadDir(oldCloudConfigDir)
		if err != nil {
			return fmt.Errorf("reading legacy .bbl dir contents: %s", err)
		}

		for _, file := range files {
			oldFile := filepath.Join(oldCloudConfigDir, file.Name())
			oldFileContent, err := m.fileIO.ReadFile(oldFile)
			if err != nil {
				return fmt.Errorf("reading %s: %s", oldFile, err)
			}

			newFile := filepath.Join(cloudConfigDir, file.Name())
			err = m.fileIO.WriteFile(newFile, oldFileContent, StateMode)
			if err != nil {
				return fmt.Errorf("migrating %s to %s: %s", oldFile, newFile, err)
			}
		}

		err = m.fileIO.RemoveAll(m.store.GetOldBblDir())
		if err != nil {
			return fmt.Errorf("removing legacy .bbl dir: %s", err)
		}
	}
	return nil
}

func (m Migrator) MigrateTerraformVars(varsDir string) error {
	tfVarsPath := filepath.Join(varsDir, "terraform.tfvars")
	bblVarsPath := filepath.Join(varsDir, "bbl.tfvars")
	if _, err := m.fileIO.Stat(tfVarsPath); err == nil {
		err = m.fileIO.Rename(tfVarsPath, bblVarsPath)
		if err != nil {
			return fmt.Errorf("migrating tfvars: %s", err)
		}
	}
	return nil
}

func (m Migrator) burnAfterReadingLegacyVarsStore(varsDir, deployment string) (string, error) {
	legacyVarsStore := filepath.Join(varsDir, fmt.Sprintf("%s-variables.yml", deployment))
	if _, err := m.fileIO.Stat(legacyVarsStore); err == nil {
		boshVars, err := m.fileIO.ReadFile(legacyVarsStore)
		if err != nil {
			return "", fmt.Errorf("reading legacy %s vars store: %s", deployment, err)
		}
		if err := m.fileIO.Remove(legacyVarsStore); err != nil {
			return "", fmt.Errorf("removing legacy %s vars store: %s", deployment, err) //not tested
		}

		return string(boshVars), nil
	} else {
		return "", nil
	}
}

func (m Migrator) MigrateDirectorVars(state State, varsDir string) (State, error) {
	err := m.migrateVarsStore(state.BOSH.Variables, "director", varsDir)
	if err != nil {
		return State{}, err
	}
	state.BOSH.Variables = ""
	return state, nil
}

func (m Migrator) MigrateJumpboxVars(state State, varsDir string) (State, error) {
	err := m.migrateVarsStore(state.Jumpbox.Variables, "jumpbox", varsDir)
	if err != nil {
		return State{}, err
	}
	state.Jumpbox.Variables = ""
	return state, nil
}

func (m Migrator) migrateVarsStore(variables, deployment, varsDir string) error {
	boshVars, err := m.burnAfterReadingLegacyVarsStore(varsDir, deployment)
	if err != nil {
		return err
	}

	if variables == "" {
		variables = boshVars
	}

	if variables != "" {
		err := m.fileIO.WriteFile(filepath.Join(varsDir, fmt.Sprintf("%s-vars-store.yml", deployment)), []byte(variables), StateMode)
		if err != nil {
			return fmt.Errorf("migrating %s variables: %s", deployment, err)
		}
	}

	return nil
}
