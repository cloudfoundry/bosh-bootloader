package acceptance_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Upgrade", func() {
	var (
		oldBBL  actors.BBL
		newBBL  actors.BBL
		boshcli actors.BOSHCLI

		sshSession *gexec.Session
		f          *os.File
	)

	BeforeEach(func() {
		acceptance.SkipUnless("upgrade")

		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		var bblBinaryLocation string
		if runtime.GOOS == "darwin" {
			bblBinaryLocation = "https://github.com/cloudfoundry/bosh-bootloader/releases/download/v5.11.5/bbl-v5.11.5_osx"
		} else {
			bblBinaryLocation = "https://github.com/cloudfoundry/bosh-bootloader/releases/download/v5.11.5/bbl-v5.11.5_linux_x86-64"
		}

		resp, err := http.Get(bblBinaryLocation)
		Expect(err).NotTo(HaveOccurred())

		f, err = ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())

		_, err = io.Copy(f, resp.Body)
		Expect(err).NotTo(HaveOccurred())

		err = os.Chmod(f.Name(), 0700)
		Expect(err).NotTo(HaveOccurred())

		err = f.Close()
		Expect(err).NotTo(HaveOccurred())

		envName := "upgrade-env"
		testName := os.Getenv("RUN_TEST")
		if testName != "" {
			envName = testName
		}
		oldBBL = actors.NewBBL(configuration.StateFileDir, f.Name(), configuration, envName, false)
		newBBL = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, envName, false)
		boshcli = actors.NewBOSHCLI()
	})

	AfterEach(func() {
		acceptance.SkipUnless("upgrade")

		if sshSession != nil {
			sshSession.Interrupt()
			Eventually(sshSession, "5s").Should(gexec.Exit())
		}

		By("trying to destroy with the old bbl", func() {
			session := oldBBL.Destroy()
			Eventually(session, bblDownTimeout).Should(gexec.Exit())
		})

		By("trying to destroy with the latest bbl", func() {
			session := newBBL.Destroy()
			Eventually(session, bblDownTimeout).Should(gexec.Exit())
		})

		err := os.Remove(f.Name())
		Expect(err).NotTo(HaveOccurred())
	})

	It("is able to upgrade from an environment bbl'd up with an older version of bbl", func() {
		By("cleaning up any leftovers", func() {
			session := newBBL.CleanupLeftovers(newBBL.PredefinedEnvID())
			Eventually(session, bblLeftoversTimeout).Should(gexec.Exit())
		})

		By("bbl'ing up with old bbl", func() {
			session := oldBBL.Up("--name", oldBBL.PredefinedEnvID())
			Eventually(session, bblUpTimeout).Should(gexec.Exit(0))
		})

		By("verifying the director has a private ip", func() {
			Expect(oldBBL.DirectorAddress()).To(Equal("https://10.0.0.6:25555"))
		})

		By("starting an ssh tunnel to talk to the director", func() {
			sshSession = oldBBL.StartSSHTunnel()
		})

		By("verifying the director exists", func() {
			exists, err := boshcli.DirectorExists(oldBBL.DirectorAddress(), oldBBL.DirectorUsername(), oldBBL.DirectorPassword(), oldBBL.SaveDirectorCA())
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		By("upgrading to the latest bbl", func() {
			session := newBBL.Plan()
			Eventually(session, bblPlanTimeout).Should(gexec.Exit(0))

			session = newBBL.Up()
			Eventually(session, bblUpTimeout).Should(gexec.Exit(0))
		})

		By("exporting BOSH_ALL_PROXY to talk to the director", func() {
			newBBL.ExportBoshAllProxy()
		})

		By("verifying the director still exists", func() {
			exists, err := boshcli.DirectorExists(newBBL.DirectorAddress(), newBBL.DirectorUsername(), newBBL.DirectorPassword(), newBBL.SaveDirectorCA())
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})
	})
})
