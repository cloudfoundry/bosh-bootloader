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

var _ = Describe("Stack Migration", func() {
	var (
		bblStack     actors.BBL
		bblTerraform actors.BBL
		aws          actors.AWS
		boshcli      actors.BOSHCLI
		state        acceptance.State

		f *os.File
	)

	BeforeEach(func() {
		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		var bblBinaryLocation string
		if runtime.GOOS == "darwin" {
			bblBinaryLocation = "https://www.github.com/cloudfoundry/bosh-bootloader/releases/download/v3.2.4/bbl-v3.2.4_osx"
		} else {
			bblBinaryLocation = "https://www.github.com/cloudfoundry/bosh-bootloader/releases/download/v3.2.4/bbl-v3.2.4_linux_x86-64"
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

		envName := "stack-migration-env"
		testName := os.Getenv("RUN_TEST")
		if testName != "" {
			envName = testName
		}
		bblStack = actors.NewBBL(configuration.StateFileDir, f.Name(), configuration, envName)
		bblTerraform = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, envName)
		aws = actors.NewAWS(configuration)
		boshcli = actors.NewBOSHCLI()
		state = acceptance.NewState(configuration.StateFileDir)
	})

	AfterEach(func() {
		By("destroying with the old bbl", func() {
			session := bblStack.Destroy()
			Eventually(session, 10*time.Minute).Should(gexec.Exit())
		})

		By("destroying with the latest bbl", func() {
			session := bblTerraform.Destroy()
			Eventually(session, 10*time.Minute).Should(gexec.Exit())
		})

		err := os.Remove(f.Name())
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Up", func() {
		It("is able to bbl up idempotently with a director", func() {
			acceptance.SkipUnless("stack-migration-up")
			var (
				stackName       string
				directorAddress string
				caCertPath      string
			)

			By("bbl'ing up with cloudformation", func() {
				session := bblStack.Up("--iaas", "aws", "--name", bblStack.PredefinedEnvID())
				Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
			})

			By("verifying the stack exists", func() {
				stackName = state.StackName()
				Expect(aws.StackExists(stackName)).To(BeTrue())
			})

			By("verifying the director exists", func() {
				directorAddress = bblStack.DirectorAddress()
				caCertPath = bblStack.SaveDirectorCA()

				exists, err := boshcli.DirectorExists(directorAddress, caCertPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())
			})

			By("migrating to terraform with latest bbl", func() {
				session := bblTerraform.Up()
				Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
			})

			By("verifying the stack doesn't exists", func() {
				Expect(aws.StackExists(stackName)).To(BeFalse())
			})

			By("verifying the director still exists", func() {
				exists, err := boshcli.DirectorExists(directorAddress, caCertPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())
			})
		})
	})
})
