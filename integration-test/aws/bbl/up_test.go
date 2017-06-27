package integration_test

import (
	"fmt"
	"net/url"

	"golang.org/x/crypto/ssh"

	integration "github.com/cloudfoundry/bosh-bootloader/integration-test"
	"github.com/cloudfoundry/bosh-bootloader/integration-test/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("up test", func() {
	var (
		bbl     actors.BBL
		aws     actors.AWS
		bosh    actors.BOSH
		boshcli actors.BOSHCLI
		state   integration.State
	)

	BeforeEach(func() {
		var err error
		configuration, err := integration.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "up-env")
		aws = actors.NewAWS(configuration)
		bosh = actors.NewBOSH()
		boshcli = actors.NewBOSHCLI()
		state = integration.NewState(configuration.StateFileDir)

		bbl.Up(actors.AWSIAAS, []string{"--name", bbl.PredefinedEnvID()})
	})

	AfterEach(func() {
		if !CurrentGinkgoTestDescription().Failed {
			bbl.Destroy()
		}
	})

	It("bbl's up a new bosh director", func() {
		By("checking if the bosh director exists", func() {
			directorAddress := bbl.DirectorAddress()
			caCertPath := bbl.SaveDirectorCA()

			exists, err := boshcli.DirectorExists(directorAddress, caCertPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		By("checking if ssh'ing works", func() {
			privateKey, err := ssh.ParsePrivateKey([]byte(bbl.SSHKey()))
			Expect(err).NotTo(HaveOccurred())

			directorAddressURL, err := url.Parse(bbl.DirectorAddress())
			Expect(err).NotTo(HaveOccurred())

			address := fmt.Sprintf("%s:22", directorAddressURL.Hostname())
			_, err = ssh.Dial("tcp", address, &ssh.ClientConfig{
				User: "jumpbox",
				Auth: []ssh.AuthMethod{
					ssh.PublicKeys(privateKey),
				},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			})
			Expect(err).NotTo(HaveOccurred())
		})

		By("checking if the instances exist", func() {
			instances := aws.Instances(fmt.Sprintf("%s-vpc", bbl.PredefinedEnvID()))
			Expect(instances).To(HaveLen(2))
			Expect(instances).To(ConsistOf([]string{"bosh/0", fmt.Sprintf("%s-nat", bbl.PredefinedEnvID())}))

			tags := aws.GetEC2InstanceTags(fmt.Sprintf("%s-nat", bbl.PredefinedEnvID()))
			Expect(tags["EnvID"]).To(Equal(bbl.PredefinedEnvID()))
		})
	})
})
