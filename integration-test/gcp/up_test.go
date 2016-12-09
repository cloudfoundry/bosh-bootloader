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

		By("creating a load balancer", func() {
			bbl.CreateGCPLB("concourse")
		})

		By("checking that the target pool exists", func() {
			targetPool, err := gcp.GetTargetPool(envID + "-concourse")
			Expect(err).NotTo(HaveOccurred())
			Expect(targetPool.Name).NotTo(BeNil())
			Expect(targetPool.Name).To(Equal(envID + "-concourse"))
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

		By("checking that the health service monitor does not exist", func() {
			healthCheck, _ := gcp.GetHealthCheck(envID + "-concourse")
			Expect(healthCheck).To(BeNil())
		})

		By("checking that the target pool does not exists", func() {
			targetPool, _ := gcp.GetTargetPool(envID + "-concourse")
			Expect(targetPool).To(BeNil())
		})
	})
})
