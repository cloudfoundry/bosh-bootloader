package integration_test

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bbl", func() {
	var (
		tempDirectory        string
		stackManager         cloudformation.StackManager
		cloudFormationClient cloudformation.Client
		state                storage.State
	)

	BeforeEach(func() {
		var err error
		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		stackManager = cloudformation.NewStackManager(application.NewLogger(os.Stdout))
		cloudFormationClient, err = aws.NewClientProvider().CloudFormationClient(aws.Config{
			AccessKeyID:     config.AWSAccessKeyID,
			SecretAccessKey: config.AWSSecretAccessKey,
			Region:          config.AWSRegion,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			session := bbl([]string{
				"--aws-access-key-id", config.AWSAccessKeyID,
				"--aws-secret-access-key", config.AWSSecretAccessKey,
				"--aws-region", config.AWSRegion,
				"--state-dir", tempDirectory,
				"destroy",
				"--no-confirm",
			})
			Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
		}
	})

	It("applies an lb to the stack", func() {
		By("running bbl up", func() {
			session := bbl([]string{
				"--aws-access-key-id", config.AWSAccessKeyID,
				"--aws-secret-access-key", config.AWSSecretAccessKey,
				"--aws-region", config.AWSRegion,
				"--state-dir", tempDirectory,
				"unsupported-deploy-bosh-on-aws-for-concourse",
			})
			Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
		})

		By("checking that there is no lb", func() {
			var err error
			state, err = loadStateJson(tempDirectory)
			Expect(err).NotTo(HaveOccurred())

			stack, err := stackManager.Describe(cloudFormationClient, state.Stack.Name)
			Expect(err).NotTo(HaveOccurred())

			Expect(stack.Outputs).NotTo(HaveKey("LB"))
			Expect(stack.Outputs).NotTo(HaveKey("LBURL"))
		})

		By("running bbl up --lb-type=concourse", func() {
			session := bbl([]string{
				"--aws-access-key-id", config.AWSAccessKeyID,
				"--aws-secret-access-key", config.AWSSecretAccessKey,
				"--aws-region", config.AWSRegion,
				"--state-dir", tempDirectory,
				"unsupported-deploy-bosh-on-aws-for-concourse",
				"--lb-type", "concourse",
			})
			Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
		})

		By("checking that there is a load balancer for concourse", func() {
			stack, err := stackManager.Describe(cloudFormationClient, state.Stack.Name)
			Expect(err).NotTo(HaveOccurred())

			Expect(stack.Outputs["LB"]).To(ContainSubstring("Concours"))
			Expect(stack.Outputs["LBURL"]).To(ContainSubstring("Concours"))
		})

		By("running bbl up", func() {
			session := bbl([]string{
				"--aws-access-key-id", config.AWSAccessKeyID,
				"--aws-secret-access-key", config.AWSSecretAccessKey,
				"--aws-region", config.AWSRegion,
				"--state-dir", tempDirectory,
				"unsupported-deploy-bosh-on-aws-for-concourse",
			})
			Eventually(session, 5*time.Minute).Should(gexec.Exit(0))
		})

		By("checking that the concourse load balancer is still there", func() {
			stack, err := stackManager.Describe(cloudFormationClient, state.Stack.Name)
			Expect(err).NotTo(HaveOccurred())

			Expect(stack.Outputs["LB"]).To(ContainSubstring("Concours"))
			Expect(stack.Outputs["LBURL"]).To(ContainSubstring("Concours"))
		})

		By("destroying bbl", func() {
			session := bbl([]string{
				"--aws-access-key-id", config.AWSAccessKeyID,
				"--aws-secret-access-key", config.AWSSecretAccessKey,
				"--aws-region", config.AWSRegion,
				"--state-dir", tempDirectory,
				"destroy",
				"--no-confirm",
			})
			Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
		})
	})
})
