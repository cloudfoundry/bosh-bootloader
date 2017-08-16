package acceptance_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"time"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"

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
		var err error
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

		bblStack = actors.NewBBL(configuration.StateFileDir, f.Name(), configuration, "stack-migration-env")
		bblTerraform = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "stack-migration-env")
		aws = actors.NewAWS(configuration)
		boshcli = actors.NewBOSHCLI()
		state = acceptance.NewState(configuration.StateFileDir)
	})

	AfterEach(func() {
		session := bblStack.Destroy()
		Eventually(session, 10*time.Minute).Should(gexec.Exit())

		session = bblTerraform.Destroy()
		Eventually(session, 10*time.Minute).Should(gexec.Exit())

		err := os.Remove(f.Name())
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Up", func() {
		It("is able to bbl up idempotently with a director", func() {
			var (
				stackName       string
				directorAddress string
				caCertPath      string
			)

			By("bbl'ing up with cloudformation", func() {
				session := bblStack.Up("aws", []string{"--name", bblStack.PredefinedEnvID()})
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
				session := bblTerraform.Up("aws", []string{})
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

	Describe("Create LBs", func() {
		It("is able to bbl create-lbs", func() {
			var (
				stackName string
				lbNames   []string
			)

			By("bbl'ing up with cloudformation", func() {
				session := bblStack.Up("aws", []string{"--name", bblStack.PredefinedEnvID()})
				Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
			})

			By("verifying the stack exists", func() {
				stackName = state.StackName()
				Expect(aws.StackExists(stackName)).To(BeTrue())
			})

			By("verifying there are no LBs", func() {
				lbNames = aws.LoadBalancers(fmt.Sprintf("vpc-%s", bblStack.PredefinedEnvID()))
				Expect(lbNames).To(BeEmpty())
			})

			By("creating a concourse load balancer", func() {
				certPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
				Expect(err).NotTo(HaveOccurred())

				chainPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CHAIN)
				Expect(err).NotTo(HaveOccurred())

				keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
				Expect(err).NotTo(HaveOccurred())

				session := bblTerraform.CreateLB("concourse", certPath, keyPath, chainPath)
				Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
			})

			By("verifying that no stack exists", func() {
				Expect(aws.StackExists(stackName)).To(BeFalse())
			})

			By("checking that the LB was created", func() {
				vpcName := fmt.Sprintf("%s-vpc", bblStack.PredefinedEnvID())
				Expect(aws.LoadBalancers(vpcName)).To(HaveLen(1))
				Expect(aws.LoadBalancers(vpcName)).To(ConsistOf(
					MatchRegexp(".*-concourse-lb"),
				))
			})
		})

		It("deletes lbs from older bbl", func() {
			var (
				stackName string
				lbNames   []string
			)

			By("bbl'ing up with cloudformation", func() {
				session := bblStack.Up("aws", []string{"--name", bblStack.PredefinedEnvID()})
				Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
			})

			By("verifying the stack exists", func() {
				stackName = state.StackName()
				Expect(aws.StackExists(stackName)).To(BeTrue())
			})

			By("verifying there are no LBs", func() {
				lbNames = aws.LoadBalancers(fmt.Sprintf("vpc-%s", bblStack.PredefinedEnvID()))
				Expect(lbNames).To(BeEmpty())
			})

			By("creating cf lbs", func() {
				certPath, err := testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_CERT)
				Expect(err).NotTo(HaveOccurred())

				chainPath, err := testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_CHAIN)
				Expect(err).NotTo(HaveOccurred())

				keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_KEY)
				Expect(err).NotTo(HaveOccurred())

				session := bblStack.CreateLB("cf", certPath, keyPath, chainPath)
				Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
			})

			By("checking that the LB was created", func() {
				vpcName := fmt.Sprintf("vpc-%s", bblStack.PredefinedEnvID())
				Expect(aws.LoadBalancers(vpcName)).To(HaveLen(2))
				Expect(aws.LoadBalancers(vpcName)).To(ConsistOf(
					MatchRegexp("stack-.*-CFSSHPro-.*"),
					MatchRegexp("stack-.*-CFRouter-.*"),
				))
			})

			By("deleting the LBs", func() {
				session := bblTerraform.DeleteLBs()
				Eventually(session, 15*time.Minute).Should(gexec.Exit(0))
			})

			By("verifying that no stack exists", func() {
				Expect(aws.StackExists(stackName)).To(BeFalse())
			})

			By("confirming that the cf lbs do not exist", func() {
				vpcName := fmt.Sprintf("%s-vpc", bblStack.PredefinedEnvID())
				Expect(aws.LoadBalancers(vpcName)).To(BeEmpty())
			})
		})
	})
})
