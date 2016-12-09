package integration_test

import (
	"fmt"

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

	It("successfully bbls up and destroys", func() {
		var (
			expectedSSHKey  string
			envID           string
			directorAddress string
			caCertPath      string
		)

		By("calling bbl up", func() {
			bbl.Up(actors.GCPIAAS)

			envID = state.EnvID()
		})

		By("checking the ssh key exists", func() {
			expectedSSHKey = fmt.Sprintf("vcap:%s vcap", state.SSHPublicKey())

			actualSSHKeys, err := gcp.SSHKey()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualSSHKeys).To(ContainSubstring(expectedSSHKey))
		})

		By("checking the network and subnet", func() {
			network, err := gcp.GetNetwork(envID + "-network")
			Expect(err).NotTo(HaveOccurred())
			Expect(network).NotTo(BeNil())

			subnet, err := gcp.GetSubnet(envID + "-subnet")
			Expect(err).NotTo(HaveOccurred())
			Expect(subnet).NotTo(BeNil())
		})

		By("checking the static ip", func() {
			address, err := gcp.GetAddress(envID + "-bosh-external-ip")
			Expect(err).NotTo(HaveOccurred())
			Expect(address).NotTo(BeNil())
		})

		By("checking the open and internal firewall rules", func() {
			boshOpenFirewallRule, err := gcp.GetFirewallRule(envID + "-bosh-open")
			Expect(err).NotTo(HaveOccurred())
			Expect(boshOpenFirewallRule).NotTo(BeNil())

			internalFirewallRule, err := gcp.GetFirewallRule(envID + "-internal")
			Expect(err).NotTo(HaveOccurred())
			Expect(internalFirewallRule).NotTo(BeNil())
		})

		By("checking that the bosh director exists", func() {
			directorAddress = bbl.DirectorAddress()
			caCertPath = bbl.SaveDirectorCA()
			exists, err := boshcli.DirectorExists(directorAddress, caCertPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		By("checking that the cloud config exists", func() {
			directorUsername := bbl.DirectorUsername()
			directorPassword := bbl.DirectorPassword()

			cloudConfig, err := boshcli.CloudConfig(directorAddress, caCertPath, directorUsername, directorPassword)
			Expect(err).NotTo(HaveOccurred())
			Expect(cloudConfig).NotTo(BeEmpty())
		})

		By("calling bbl destroy", func() {
			bbl.Destroy()
		})

		By("checking the ssh key does not exist", func() {
			actualSSHKeys, err := gcp.SSHKey()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualSSHKeys).NotTo(ContainSubstring(expectedSSHKey))
		})

		By("checking the network and subnet do not exist", func() {
			network, _ := gcp.GetNetwork(envID + "-network")
			Expect(network).To(BeNil())

			subnet, _ := gcp.GetSubnet(envID + "-subnet")
			Expect(subnet).To(BeNil())
		})

		By("checking the static ip does not exist", func() {
			address, _ := gcp.GetAddress(envID + "-bosh-external-ip")
			Expect(address).To(BeNil())
		})

		By("checking the open and internal firewall rules do not exist", func() {
			boshOpenFirewallRule, _ := gcp.GetFirewallRule(envID + "-bosh-open")
			Expect(boshOpenFirewallRule).To(BeNil())

			internalFirewallRule, _ := gcp.GetFirewallRule(envID + "-internal")
			Expect(internalFirewallRule).To(BeNil())
		})

		By("checking that the bosh director does not exists", func() {
			exists, _ := boshcli.DirectorExists(directorAddress, caCertPath)
			Expect(exists).To(BeFalse())
		})
	})
})
