package commands_test

import (
	"os"

	"github.com/cloudfoundry/bosh-bootloader/fakes"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("RotateCA", func() {
	Describe("Implementation", func() {
		It("should implement Command", func() {
			var rotateCA commands.Command = commands.RotateCA{}
			rotateCA.Usage()
		})
	})

	Describe("Execute", func() {
		var (
			rotateCA commands.RotateCA
			subcommandFlags []string
			state storage.State
			boshCLI *fakes.BOSHCLI
			store *fakes.StateStore
		)

		BeforeEach(func() {
			state = storage.State{Jumpbox: storage.Jumpbox{URL: "some-jumpbox"}}
			boshCLI = &fakes.BOSHCLI{}
			store = &fakes.StateStore{}
			
			store.GetStateDirCall.Returns.Directory = "some-state-dir"
			store.GetCall.Returns.State.IAAS = "some-iaas"
			boshCLI.RunCall.Returns.Error = nil

			rotateCA.BOSH = boshCLI
			rotateCA.StateDirSource = store
		})
		
		It("should call bosh create-env with the ops file", func() {
			err := rotateCA.Execute(subcommandFlags, state)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshCLI.RunCall.CallCount).To(Equal(1))
			Expect(boshCLI.RunCall.Receives.WorkingDirectory).To(Equal("some-state-dir"))
			Expect(boshCLI.RunCall.Receives.Args).To(Equal([]string{
				"create-env",
				"bosh-deployment/bosh.yml",
				"-o", "bosh-deployment/add-new-ca.yml",
				"-o", "bosh-deployment/some-iaas/cpi.yml",
				"--vars-store", "director-vars-store.yml",
				"--vars-file", "director-vars-file.yml",
			}))
		})

		It("should return an error when there is no director", func() {
			state.Jumpbox.URL = ""

			err := rotateCA.Execute(subcommandFlags, state)
			Expect(err).To(MatchError("no jumpbox"))
		})

		
	})

	Describe("WriteAddNewCA", func() {
		Context("when called", func() {
			var rotateCA commands.RotateCA

			fileIO := &fakes.FileIO{}
			store := &fakes.StateStore{}
			store.GetStateDirCall.Returns.Directory = "some-state-dir"

			rotateCA.FileSystem = fileIO
			rotateCA.StateDirSource = store

			rotateCA.WriteAddNewCA()

			It("should write a file", func() {
				Expect(fileIO.WriteFileCall.CallCount).To(Equal(1))
				Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal("some-state-dir/bosh-deployment/add-new-ca.yml"))
				Expect(fileIO.WriteFileCall.Receives[0].Contents).To(Equal([]byte(addNewCAFileContents)))
				Expect(fileIO.WriteFileCall.Receives[0].Mode).To(Equal(os.FileMode(0644)))
			})

		})
	})
})

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
