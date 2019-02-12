package commands

import (
	"errors"
	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"io"
	"path/filepath"
)

// type fs interface {

// 	fileio.Remover
// 	fileio.AllRemover
// 	fileio.Stater
// 	fileio.AllMkdirer
// 	fileio.DirReader
// }

type Runner interface {
	Run(stdout io.Writer, workingDirectory string, args []string) error
}

type RotateCA struct {
	FileSystem     fileio.FileWriter
	BOSH           Runner
	StateDirSource StateDirGetter
}

func (cmd RotateCA) CheckFastFails(subcommandFlags []string, state storage.State) error {
	return nil
}

func (cmd RotateCA) Usage() string {
	return ""
}

func (cmd RotateCA) Execute(subcommandFlags []string, state storage.State) error {
	if state.Jumpbox.URL == "" {
		return errors.New("no jumpbox")
	}
	cmd.BOSH.Run(nil, cmd.StateDirSource.GetStateDir(), []string{
		"create-env", "bosh-deployment/bosh.yml",
		
		"-o", "bosh-deployment/add-new-ca.yml",
		"-o", "bosh-deployment/some-iaas/cpi.yml",
		"--vars-store", "director-vars-store.yml",
		"--vars-file", "director-vars-file.yml",
	})
	return nil
}

type StateDirGetter interface {
	GetStateDir() string
}

func (cmd RotateCA) WriteAddNewCA() error {
	var fp string
	fp = filepath.Join(cmd.StateDirSource.GetStateDir(), "bosh-deployment", "add-new-ca.yml")
	err := cmd.FileSystem.WriteFile(fp, []byte(addNewCAFileContents), 0644)
	if err != nil {
		return err
	}
	return nil
}

const addNewCAFileContents = `---
- type: replace
  path: /instance_groups/name=bosh/properties/nats/tls/ca?
  value: ((nats_server_tls.ca))((nats_server_tls_2.ca))

- type: replace
  path: /instance_groups/name=bosh/properties/nats/tls/client_ca?
  value:
    certificate: ((nats_ca_2.certificate))
    private_key: ((nats_ca_2.private_key))

- type: replace
  path: /instance_groups/name=bosh/properties/nats/tls/director?
  value:
    certificate: ((nats_clients_director_tls_2.certificate))
    private_key: ((nats_clients_director_tls_2.private_key))

- type: replace
  path: /instance_groups/name=bosh/properties/nats/tls/health_monitor?
  value:
    certificate: ((nats_clients_health_monitor_tls_2.certificate))
    private_key: ((nats_clients_health_monitor_tls_2.private_key))

- type: replace
  path: /variables/-
  value:
    name: nats_ca_2
    type: certificate
    options:
      is_ca: true
      common_name: default.nats-ca.bosh-internal

- type: replace
  path: /variables/-
  value:
    name: nats_server_tls_2
    type: certificate
    options:
      ca: nats_ca_2
      common_name: default.nats.bosh-internal
      alternative_names: [((internal_ip))]
      extended_key_usage:
      - server_auth

- type: replace
  path: /variables/-
  value:
    name: nats_clients_director_tls_2
    type: certificate
    options:
      ca: nats_ca_2
      common_name: default.director.bosh-internal
      extended_key_usage:
      - client_auth

- type: replace
  path: /variables/-
  value:
    name: nats_clients_health_monitor_tls_2
    type: certificate
    options:
      ca: nats_ca_2
      common_name: default.hm.bosh-internal
      extended_key_usage:
	  - client_auth
`
