package acceptance_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"time"

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
		state   acceptance.State
		f       *os.File
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
		oldBBL = actors.NewBBL(configuration.StateFileDir, f.Name(), configuration, envName)
		newBBL = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, envName)
		boshcli = actors.NewBOSHCLI()
		state = acceptance.NewState(configuration.StateFileDir)
	})

	AfterEach(func() {
		acceptance.SkipUnless("upgrade")

		By("destroying with the old bbl", func() {
			session := oldBBL.Destroy()
			Eventually(session, 10*time.Minute).Should(gexec.Exit())
		})

		By("destroying with the latest bbl", func() {
			session := newBBL.Destroy()
			Eventually(session, 10*time.Minute).Should(gexec.Exit())
		})

		err := os.Remove(f.Name())
		Expect(err).NotTo(HaveOccurred())
	})

	It("is able to upgrade from an environment bbl'd up with an older version of bbl", func() {
		By("cleaning up any leftovers", func() {
			session := newBBL.CleanupLeftovers(newBBL.PredefinedEnvID())
			Eventually(session, 10*time.Minute).Should(gexec.Exit())
		})

		By("bbl'ing up with old bbl", func() {
			session := oldBBL.Up("--name", oldBBL.PredefinedEnvID())
			Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
		})

		By("verifying the director has a private ip", func() {
			Expect(oldBBL.DirectorAddress()).To(Equal("https://10.0.0.6:25555"))
		})

		By("verifying the director exists", func() {
			exists, err := boshcli.DirectorExists(oldBBL.DirectorAddress(), oldBBL.DirectorUsername(), oldBBL.DirectorPassword(), oldBBL.SaveDirectorCA())
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		By("upgrading to the latest bbl", func() {
			session := newBBL.Up()
			Eventually(session, 60*time.Minute).Should(gexec.Exit(0))
		})

		By("exporting environment variables to talk to the director", func() {
			newBBL.ExportBoshAllProxy()
		})

		directorAddress := newBBL.DirectorAddress()

		By("verifying the director has the same private ip", func() {
			Expect(directorAddress).To(Equal("https://10.0.0.6:25555"))
		})

		By("verifying the director still exists", func() {
			exists, err := boshcli.DirectorExists(directorAddress, newBBL.DirectorUsername(), newBBL.DirectorPassword(), newBBL.SaveDirectorCA())
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})
	})
})
