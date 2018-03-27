package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/cloudfoundry/bosh-bootloader/fileio"
	uuid "github.com/nu7hatch/gouuid"
)

var (
	marshalIndent = json.MarshalIndent
	uuidNewV4     = uuid.NewV4
)

const (
	STATE_SCHEMA = 14

	OS_READ_WRITE_MODE = os.FileMode(0644)
	StateFileName      = "bbl-state.json"
)

type Store struct {
	dir              string
	fs               fs
	garbageCollector garbageCollector
	stateSchema      int
}

type fs interface {
	fileio.FileWriter
	fileio.Remover
	fileio.AllRemover
	fileio.Stater
	fileio.AllMkdirer
	fileio.DirReader
}

type garbageCollector interface {
	Remove(d string) error
}

func NewStore(dir string, fs fs, garbageCollector garbageCollector) Store {
	return Store{
		dir:              dir,
		fs:               fs,
		garbageCollector: garbageCollector,
		stateSchema:      STATE_SCHEMA,
	}
}

func (s Store) Set(state State) error {
	_, err := s.fs.Stat(s.dir)
	if err != nil {
		return fmt.Errorf("Stat state dir: %s", err)
	}

	if reflect.DeepEqual(state, State{}) {
		err := s.garbageCollector.Remove(s.dir)
		if err != nil {
			return fmt.Errorf("Garbage collector clean up: %s", err)
		}
		return nil
	}

	state.Version = s.stateSchema

	if state.ID == "" {
		uuid, err := uuidNewV4()
		if err != nil {
			return fmt.Errorf("Create state ID: %s", err)
		}
		state.ID = uuid.String()
	}

	jsonData, err := marshalIndent(state, "", "\t")
	if err != nil {
		return err
	}

	stateFile := filepath.Join(s.dir, StateFileName)
	err = s.fs.WriteFile(stateFile, jsonData, os.FileMode(0644))
	if err != nil {
		return err
	}

	return nil
}

func (s Store) GetStateDir() string {
	return s.dir
}

func (s Store) GetCloudConfigDir() (string, error) {
	return s.getDir("cloud-config")
}

func (s Store) GetTerraformDir() (string, error) {
	return s.getDir("terraform")
}

func (s Store) GetVarsDir() (string, error) {
	return s.getDir("vars")
}

func (s Store) GetDirectorDeploymentDir() (string, error) {
	return s.getDir("bosh-deployment")
}

func (s Store) GetJumpboxDeploymentDir() (string, error) {
	return s.getDir("jumpbox-deployment")
}

func (s Store) GetOldBblDir() string {
	return filepath.Join(s.dir, ".bbl")
}

func (s Store) getDir(name string) (string, error) {
	dir := filepath.Join(s.dir, name)
	err := s.fs.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("Get %s dir: %s", name, err)
	}
	return dir, nil
}
