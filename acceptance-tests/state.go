package acceptance

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"

	. "github.com/onsi/gomega"
)

type State struct {
	stateFilePath string
}

type state struct {
	EnvID   string `json:"envID"`
	TFState string `json:"tfState"`
	BOSH    struct {
		State    map[string]interface{} `json:"state"`
		Manifest string                 `json:"manifest"`
	} `json:"bosh"`
}

func (s State) Checksum() string {
	buf, err := os.ReadFile(s.stateFilePath)
	Expect(err).NotTo(HaveOccurred())
	return fmt.Sprintf("%x", md5.Sum(buf))
}

func (s State) EnvID() string {
	state := s.readStateFile()
	return state.EnvID
}

func (s State) TFState() string {
	state := s.readStateFile()
	return state.TFState
}

func (s State) BOSHState() string {
	state := s.readStateFile()
	boshState, err := json.Marshal(state.BOSH.State)
	if err != nil {
		panic(err)
	}
	return string(boshState)
}

func (s State) BOSHManifest() string {
	state := s.readStateFile()

	return state.BOSH.Manifest
}

func (s State) readStateFile() state {
	stateFile, err := os.Open(s.stateFilePath)
	Expect(err).NotTo(HaveOccurred())
	defer stateFile.Close()

	var state state
	err = json.NewDecoder(stateFile).Decode(&state)
	Expect(err).NotTo(HaveOccurred())

	return state
}
