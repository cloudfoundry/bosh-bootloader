package acceptance_test

import (
	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("lbs test", func() {
	var (
		bbl       actors.BBL
		gcp       actors.GCP
		terraform actors.Terraform
		boshcli   actors.BOSHCLI
		state     acceptance.State
	)

	BeforeEach(func() {
		var err error
		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		state = acceptance.NewState(configuration.StateFileDir)
		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "lbs-env")
		gcp = actors.NewGCP(configuration)
		terraform = actors.NewTerraform(configuration)
		boshcli = actors.NewBOSHCLI()

		bbl.Up(actors.GCPIAAS, []string{"--name", bbl.PredefinedEnvID(), "--no-director"})
	})

	AfterEach(func() {
		if !CurrentGinkgoTestDescription().Failed {
			bbl.Destroy()
		}
	})

	It("successfully creates a concourse lb", func() {
		By("creating a load balancer", func() {
			bbl.CreateLB("concourse", "", "", "")
		})

		By("confirming that target pools exist", func() {
			targetPool, err := gcp.GetTargetPool(bbl.PredefinedEnvID() + "-concourse")
			Expect(err).NotTo(HaveOccurred())
			Expect(targetPool.Name).NotTo(BeNil())
		})

		By("verifying that the bbl lbs output contains the concourse lb", func() {
			session := bbl.LBs()
			stdout := string(session.Out.Contents())
			Expect(stdout).To(MatchRegexp("Concourse LB: .*"))
		})

		By("deleting lbs", func() {
			bbl.DeleteLBs()
		})

		By("confirming that the target pools do not exist", func() {
			_, err := gcp.GetTargetPool(bbl.PredefinedEnvID() + "-concourse")
			Expect(err).To(MatchError(MatchRegexp(`The resource 'projects\/.+` + bbl.PredefinedEnvID() + "-concourse" + `' was not found`)))
		})
	})

	It("successfully creates cf lbs", func() {
		var urlToSSLCert string

		By("creating a load balancer", func() {
			certPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
			Expect(err).NotTo(HaveOccurred())

			keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
			Expect(err).NotTo(HaveOccurred())

			bbl.CreateLB("cf", certPath, keyPath, "")
		})

		By("confirming that target pools exist", func() {
			targetPools := []string{bbl.PredefinedEnvID() + "-cf-ssh-proxy", bbl.PredefinedEnvID() + "-cf-tcp-router"}
			for _, p := range targetPools {
				targetPool, err := gcp.GetTargetPool(p)
				Expect(err).NotTo(HaveOccurred())
				Expect(targetPool.Name).NotTo(BeNil())
				Expect(targetPool.Name).To(Equal(p))
			}

			targetHTTPSProxy, err := gcp.GetTargetHTTPSProxy(bbl.PredefinedEnvID() + "-https-proxy")
			Expect(err).NotTo(HaveOccurred())

			Expect(targetHTTPSProxy.SslCertificates).To(HaveLen(1))
			urlToSSLCert = targetHTTPSProxy.SslCertificates[0]
		})

		By("verifying that the bbl lbs output contains the cf lbs", func() {
			session := bbl.LBs()
			stdout := string(session.Out.Contents())
			Expect(stdout).To(MatchRegexp("CF Router LB: .*"))
			Expect(stdout).To(MatchRegexp("CF SSH Proxy LB: .*"))
			Expect(stdout).To(MatchRegexp("CF TCP Router LB: .*"))
			Expect(stdout).To(MatchRegexp("CF WebSocket LB: .*"))
		})

		By("updating the load balancer", func() {
			otherCertPath, err := testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_CERT)
			Expect(err).NotTo(HaveOccurred())

			otherKeyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_KEY)
			Expect(err).NotTo(HaveOccurred())

			bbl.UpdateLB(otherCertPath, otherKeyPath, "")
		})

		By("confirming that the cert gets updated", func() {
			targetHTTPSProxy, err := gcp.GetTargetHTTPSProxy(bbl.PredefinedEnvID() + "-https-proxy")
			Expect(err).NotTo(HaveOccurred())

			Expect(targetHTTPSProxy.SslCertificates).To(HaveLen(1))
			Expect(targetHTTPSProxy.SslCertificates[0]).NotTo(BeEmpty())
			Expect(targetHTTPSProxy.SslCertificates[0]).NotTo(Equal(urlToSSLCert))
		})

		By("deleting lbs", func() {
			bbl.DeleteLBs()
		})

		By("confirming that the target pools do not exist", func() {
			targetPools := []string{bbl.PredefinedEnvID() + "-cf-ssh-proxy", bbl.PredefinedEnvID() + "-cf-tcp-router"}
			for _, p := range targetPools {
				_, err := gcp.GetTargetPool(p)
				Expect(err).To(MatchError(MatchRegexp(`The resource 'projects\/.+` + p + `' was not found`)))
			}
		})
	})
})
