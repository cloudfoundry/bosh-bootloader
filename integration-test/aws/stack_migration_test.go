package integration_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	integration "github.com/cloudfoundry/bosh-bootloader/integration-test"
	"github.com/cloudfoundry/bosh-bootloader/integration-test/actors"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Stack Migration", func() {
	var (
		bblStack     actors.BBL
		bblTerraform actors.BBL
		aws          actors.AWS
		boshcli      actors.BOSHCLI
		state        integration.State

		f *os.File
	)

	BeforeEach(func() {
		var err error
		configuration, err := integration.LoadConfig()
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
		state = integration.NewState(configuration.StateFileDir)
	})

	AfterEach(func() {
		if !CurrentGinkgoTestDescription().Failed {
			bblTerraform.Destroy()
		}

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
				bblStack.Up(actors.AWSIAAS, []string{"--name", bblStack.PredefinedEnvID()})
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
				bblTerraform.Up(actors.AWSIAAS, []string{})
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

	PDescribe("Create LBs", func() {
		It("is able to bbl create-lbs", func() {
			var (
				stackName string
				lbNames   []string
			)

			By("bbl'ing up with cloudformation", func() {
				bblStack.Up(actors.AWSIAAS, []string{"--name", bblStack.PredefinedEnvID()})
			})

			By("verifying the stack exists", func() {
				stackName = state.StackName()
				Expect(aws.StackExists(stackName)).To(BeTrue())
			})

			By("verifying there are no LBs", func() {
				lbNames = aws.LoadBalancers()
				Expect(lbNames).To(BeEmpty())
			})

			By("creating a concourse load balancer", func() {
				certPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
				Expect(err).NotTo(HaveOccurred())

				chainPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CHAIN)
				Expect(err).NotTo(HaveOccurred())

				keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
				Expect(err).NotTo(HaveOccurred())

				bblTerraform.CreateLB("concourse", certPath, keyPath, chainPath)
			})

			By("verifying that no stack exists", func() {
				Expect(aws.StackExists(stackName)).To(BeFalse())
			})

			By("checking that the LB was created", func() {
				Expect(aws.LoadBalancers()).To(HaveLen(1))
				Expect(aws.LoadBalancers()).To(Equal([]string{"ConcourseLoadBalancer"}))
			})
		})
	})
})
