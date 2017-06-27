package integration_test

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/crypto/ssh"

	integration "github.com/cloudfoundry/bosh-bootloader/integration-test"
	"github.com/cloudfoundry/bosh-bootloader/integration-test/actors"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("load balancer tests", func() {
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

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "lbs-env")
		aws = actors.NewAWS(configuration)
		bosh = actors.NewBOSH()
		boshcli = actors.NewBOSHCLI()
		state = integration.NewState(configuration.StateFileDir)

	})

	AfterEach(func() {
		if !CurrentGinkgoTestDescription().Failed {
			bbl.DeleteLBs()
			bbl.Destroy()
		}
	})

	It("creates, updates and deletes an LB with the specified cert and key", func() {
		bbl.Up(actors.AWSIAAS, []string{"--name", bbl.PredefinedEnvID()})

		directorAddress := bbl.DirectorAddress()
		caCertPath := bbl.SaveDirectorCA()

		Expect(aws.LoadBalancers()).To(BeEmpty())

		exists, err := boshcli.DirectorExists(directorAddress, caCertPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(BeTrue())

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

		instances := aws.Instances(fmt.Sprintf("%s-vpc", bbl.PredefinedEnvID()))
		Expect(instances).To(HaveLen(2))
		Expect(instances).To(ConsistOf([]string{"bosh/0", fmt.Sprintf("%s-nat", bbl.PredefinedEnvID())}))

		tags := aws.GetEC2InstanceTags(fmt.Sprintf("%s-nat", bbl.PredefinedEnvID()))
		Expect(tags["EnvID"]).To(Equal(bbl.PredefinedEnvID()))

		certPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
		Expect(err).NotTo(HaveOccurred())

		chainPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CHAIN)
		Expect(err).NotTo(HaveOccurred())

		keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
		Expect(err).NotTo(HaveOccurred())

		otherCertPath, err := testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_CERT)
		Expect(err).NotTo(HaveOccurred())

		otherKeyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_KEY)
		Expect(err).NotTo(HaveOccurred())

		otherChainPath, err := testhelpers.WriteContentsToTempFile(testhelpers.OTHER_BBL_CHAIN)
		Expect(err).NotTo(HaveOccurred())

		bbl.CreateLB("concourse", certPath, keyPath, chainPath)

		Expect(aws.LoadBalancers()).To(HaveLen(1))
		Expect(aws.LoadBalancers()).To(Equal([]string{fmt.Sprintf("%s-concourse-lb", bbl.PredefinedEnvID())}))

		lbName := fmt.Sprintf("%s-concourse-lb", bbl.PredefinedEnvID())
		certificateName := aws.GetSSLCertificateNameByLoadBalancer(lbName)
		Expect(strings.TrimSpace(aws.DescribeCertificate(certificateName).Body)).To(Equal(strings.TrimSpace(testhelpers.BBL_CERT)))

		bbl.UpdateLB(otherCertPath, otherKeyPath, otherChainPath)
		Expect(aws.LoadBalancers()).To(HaveLen(1))
		Expect(aws.LoadBalancers()).To(Equal([]string{fmt.Sprintf("%s-concourse-lb", bbl.PredefinedEnvID())}))

		certificateName = aws.GetSSLCertificateNameByLoadBalancer(lbName)
		Expect(strings.TrimSpace(aws.DescribeCertificate(certificateName).Body)).To(Equal(strings.TrimSpace(string(testhelpers.OTHER_BBL_CERT))))

		session := bbl.LBs()
		stdout := string(session.Out.Contents())
		Expect(stdout).To(MatchRegexp("Concourse LB: .*"))
	})
})
