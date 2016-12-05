package integration_test

import (
	"fmt"
	"io/ioutil"

	integration "github.com/cloudfoundry/bosh-bootloader/integration-test"
	"github.com/cloudfoundry/bosh-bootloader/integration-test/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("up test", func() {
	var (
		bbl       actors.BBL
		gcp       actors.GCP
		terraform actors.Terraform
		boshcli   actors.BOSHCLI
		state     integration.State
	)

	BeforeEach(func() {
		var err error
		configuration, err := integration.LoadGCPConfig()
		Expect(err).NotTo(HaveOccurred())

		state = integration.NewState(configuration.StateFileDir)
		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration)
		gcp = actors.NewGCP(configuration)
		terraform = actors.NewTerraform(configuration)
		boshcli = actors.NewBOSHCLI()
	})

	AfterEach(func() {
		boshState, err := ioutil.TempFile("", "")
		if err != nil {
			fmt.Println(err)
		}

		boshManifest, err := ioutil.TempFile("", "")
		if err != nil {
			fmt.Println(err)
		}

		if _, err := boshState.Write([]byte(state.BOSHState())); err != nil {
			fmt.Println(err)
		}

		if _, err := boshManifest.Write([]byte(state.BOSHManifest())); err != nil {
			fmt.Println(err)
		}

		if err := boshcli.DeleteEnv(boshState.Name(), boshManifest.Name()); err != nil {
			fmt.Println("DeleteEnv", err)
		}

		if err := terraform.Destroy(state); err != nil {
			fmt.Println("Destroy", err)
		}

		if err := gcp.RemoveSSHKey(fmt.Sprintf("vcap:%s vcap", state.SSHPublicKey())); err != nil {
			fmt.Println("RemoveSSHKey", err)
		}
	})

	It("successfully bbls up", func() {
		bbl.Up(actors.GCPIAAS)

		By("checking the ssh key exists", func() {
			expectedSSHKey := fmt.Sprintf("vcap:%s vcap", state.SSHPublicKey())

			actualSSHKeys, err := gcp.SSHKey()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualSSHKeys).To(ContainSubstring(expectedSSHKey))
		})

		By("checking the network and subnet", func() {
			network, err := gcp.GetNetwork(state.EnvID() + "-network")
			Expect(err).NotTo(HaveOccurred())
			Expect(network).NotTo(BeNil())

			subnet, err := gcp.GetSubnet(state.EnvID() + "-subnet")
			Expect(err).NotTo(HaveOccurred())
			Expect(subnet).NotTo(BeNil())
		})

		By("checking the static ip", func() {
			address, err := gcp.GetAddress(state.EnvID() + "-bosh-external-ip")
			Expect(err).NotTo(HaveOccurred())
			Expect(address).NotTo(BeNil())
		})

		By("checking the open and internal firewall rules", func() {
			firewallRule, err := gcp.GetFirewallRule(state.EnvID() + "-bosh-open")
			Expect(err).NotTo(HaveOccurred())
			Expect(firewallRule).NotTo(BeNil())

			firewallRule, err = gcp.GetFirewallRule(state.EnvID() + "-internal")
			Expect(err).NotTo(HaveOccurred())
			Expect(firewallRule).NotTo(BeNil())
		})

		By("checking that the bosh director exists", func() {
			directorAddress := bbl.DirectorAddress()
			caCertPath := bbl.SaveDirectorCA()
			exists, err := boshcli.DirectorExists(directorAddress, caCertPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})
	})
})
