package integration_test

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

var _ = Describe("bbl", func() {
	var (
		tempDirectory        string
		stateChecksum        string
		directorURL          string
		directorUsername     string
		directorPassword     string
		state                storage.State
		stackManager         cloudformation.StackManager
		cloudFormationClient cloudformation.Client
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

	It("provisions AWS, deploys and tears down a bosh director", func() {
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

		By("getting checksum of state.json", func() {
			var err error
			stateChecksum, err = getChecksum(tempDirectory)
			Expect(err).NotTo(HaveOccurred())
		})

		By("checking that bosh director is up", func() {
			var err error
			state, err = loadStateJson(tempDirectory)
			Expect(err).NotTo(HaveOccurred())

			_, err = stackManager.Describe(cloudFormationClient, state.Stack.Name)
			Expect(err).NotTo(HaveOccurred())

			session := bbl([]string{
				"--state-dir", tempDirectory,
				"director-address",
			})
			directorURL = strings.TrimSpace(string(session.Wait().Buffer().Contents()))

			session = bbl([]string{
				"--state-dir", tempDirectory,
				"director-username",
			})
			directorUsername = strings.TrimSpace(string(session.Wait().Buffer().Contents()))

			session = bbl([]string{
				"--state-dir", tempDirectory,
				"director-password",
			})
			directorPassword = strings.TrimSpace(string(session.Wait().Buffer().Contents()))

			client := bosh.NewClient(directorURL, directorUsername, directorPassword)

			_, err = client.Info()
			Expect(err).NotTo(HaveOccurred())
		})

		By("running bbl up again", func() {
			session := bbl([]string{
				"--aws-access-key-id", config.AWSAccessKeyID,
				"--aws-secret-access-key", config.AWSSecretAccessKey,
				"--aws-region", config.AWSRegion,
				"--state-dir", tempDirectory,
				"unsupported-deploy-bosh-on-aws-for-concourse",
			})
			Eventually(session, 1*time.Minute).Should(gexec.Exit(0))
		})

		By("checking that the checksum of state.json is the same", func() {
			newChecksum, err := getChecksum(tempDirectory)
			Expect(err).NotTo(HaveOccurred())

			Expect(stateChecksum).To(Equal(newChecksum))
		})

		By("running bbl destroy", func() {
			session := bbl([]string{
				"--state-dir", tempDirectory,
				"destroy",
				"--no-confirm",
			})
			Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
		})

		By("checking that bosh director is down", func() {
			client := bosh.NewClient(directorURL, directorUsername, directorPassword)

			_, err := client.Info()
			Expect(err.(*url.Error).Op).To(Equal("Get"))
			Expect(err.(*url.Error).URL).To(Equal(fmt.Sprintf("%s/info", directorURL)))
		})

		By("checking that the stack has been deleted", func() {
			_, err := stackManager.Describe(cloudFormationClient, state.Stack.Name)
			Expect(err).To(MatchError(cloudformation.StackNotFound))
		})
	})
})

func getChecksum(stateDirectory string) (string, error) {
	buf, err := ioutil.ReadFile(filepath.Join(stateDirectory, "state.json"))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(buf)), nil
}

func loadStateJson(stateDirectory string) (storage.State, error) {
	stateFile, err := os.Open(filepath.Join(stateDirectory, "state.json"))
	if err != nil {
		return storage.State{}, err
	}
	defer stateFile.Close()

	var state storage.State
	if err := json.NewDecoder(stateFile).Decode(&state); err != nil {
		return storage.State{}, err
	}

	return state, nil
}
