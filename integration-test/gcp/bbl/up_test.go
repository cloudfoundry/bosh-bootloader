package integration_test

import (
	"fmt"

	integration "github.com/cloudfoundry/bosh-bootloader/integration-test"
	"github.com/cloudfoundry/bosh-bootloader/integration-test/actors"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"

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

		envID string
	)

	BeforeEach(func() {
		var err error
		configuration, err := integration.LoadGCPConfig()
		Expect(err).NotTo(HaveOccurred())

		envID = configuration.GCPEnvPrefix + "bbl-ci-env"
		state = integration.NewState(configuration.StateFileDir)
		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, envID)
		gcp = actors.NewGCP(configuration)
		terraform = actors.NewTerraform(configuration)
		boshcli = actors.NewBOSHCLI()
	})

	It("successfully bbls up and destroys", func() {
		var (
			expectedSSHKey  string
			directorAddress string
			caCertPath      string
			urlToSSLCert    string
		)

		By("calling bbl up", func() {
			bbl.Up(actors.GCPIAAS, []string{"--name", envID})
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
			certPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
			Expect(err).NotTo(HaveOccurred())

			keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
			Expect(err).NotTo(HaveOccurred())

			chainPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CHAIN)
			Expect(err).NotTo(HaveOccurred())

			bbl.CreateLB("cf", certPath, keyPath, chainPath)
		})

		By("confirming that target pools exists", func() {
			targetPools := []string{envID + "-cf-ssh-proxy", envID + "-cf-tcp-router"}
			for _, p := range targetPools {
				targetPool, err := gcp.GetTargetPool(p)
				Expect(err).NotTo(HaveOccurred())
				Expect(targetPool.Name).NotTo(BeNil())
				Expect(targetPool.Name).To(Equal(p))
			}

			targetHTTPSProxy, err := gcp.GetTargetHTTPSProxy(envID + "-https-proxy")
			Expect(err).NotTo(HaveOccurred())

			Expect(targetHTTPSProxy.SslCertificates).To(HaveLen(1))
			urlToSSLCert = targetHTTPSProxy.SslCertificates[0]
		})

		By("updating the load balancer", func() {
			otherCertPath, err := testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_CERT)
			Expect(err).NotTo(HaveOccurred())

			otherKeyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_KEY)
			Expect(err).NotTo(HaveOccurred())

			bbl.UpdateLB(otherCertPath, otherKeyPath)
		})

		By("confirming that the cert gets updated", func() {
			targetHTTPSProxy, err := gcp.GetTargetHTTPSProxy(envID + "-https-proxy")
			Expect(err).NotTo(HaveOccurred())

			Expect(targetHTTPSProxy.SslCertificates).To(HaveLen(1))
			Expect(targetHTTPSProxy.SslCertificates[0]).NotTo(BeEmpty())
			Expect(targetHTTPSProxy.SslCertificates[0]).NotTo(Equal(urlToSSLCert))
		})

		By("deleting lbs", func() {
			bbl.DeleteLBs()
		})

		By("confirming that the target pools do not exist", func() {
			targetPools := []string{envID + "-cf-ssh-proxy", envID + "-cf-tcp-router"}
			for _, p := range targetPools {
				_, err := gcp.GetTargetPool(p)
				Expect(err).To(MatchError(MatchRegexp(`The resource 'projects\/.+` + p + `' was not found`)))
			}
		})

		By("calling bbl destroy", func() {
			bbl.Destroy()
		})

		By("confirming the ssh key does not exist", func() {
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
			healthCheck, _ := gcp.GetHealthCheck(envID + "-cf")
			Expect(healthCheck).To(BeNil())
		})
	})
})
