package integration_test

import (
	"fmt"
	"net/url"

	integration "github.com/cloudfoundry/bosh-bootloader/integration-test"
	"github.com/cloudfoundry/bosh-bootloader/integration-test/actors"
	"golang.org/x/crypto/ssh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("up test", func() {
	var (
		bbl       actors.BBL
		gcp       actors.GCP
		terraform actors.Terraform
		boshcli   actors.BOSHCLI
		state     integration.State
	)

	BeforeEach(func() {
		var err error
		configuration, err := integration.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		state = integration.NewState(configuration.StateFileDir)
		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "jumpbox-env")
		gcp = actors.NewGCP(configuration)
		terraform = actors.NewTerraform(configuration)
		boshcli = actors.NewBOSHCLI()
	})

	AfterEach(func() {
		bbl.Destroy()
	})

	FIt("successfully bbls up and destroys", func() {
		By("calling bbl up", func() {
			bbl.Up(actors.GCPIAAS, []string{"--name", bbl.PredefinedEnvID(), "--jumpbox"})
		})

		By("checking the ssh key exists", func() {
			expectedSSHKey := fmt.Sprintf("vcap:%s vcap", state.SSHPublicKey())

			actualSSHKeys, err := gcp.SSHKey()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualSSHKeys).To(ContainSubstring(expectedSSHKey))
		})

		By("checking the user can create a tunnel", func() {
			
		})

		// bosh int ./jumpbox-vars-store.yml --path /jumpbox_ssh/private_key > jumpbox.key
		// chmod 600 jumpbox.key
		// external_ip=$(bosh int ./jumpbox-deployment-vars.yml --path /external_ip)
		// ssh -N -D 9999 jumpbox@${external_ip} -i jumpbox.key &
		// export BOSH_ALL_PROXY=socks5://localhost:9999

		// ca_cert="$(bosh int ./bosh-vars-store.yml --path /default_ca/ca)"
		// admin_password="$(bosh int ./bosh-vars-store.yml --path /admin_password)"
		// export BOSH_ENVIRONMENT=10.0.0.6
		// export BOSH_CA_CERT="${ca_cert}"
		// export BOSH_CLIENT=admin
		// export BOSH_CLIENT_SECRET=${admin_password}

		// bosh env

		By("checking that the bosh director exists", func() {
			directorAddress := bbl.DirectorAddress()
			caCertPath := bbl.SaveDirectorCA()
			exists, err := boshcli.DirectorExists(directorAddress, caCertPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		By("checking that the user can ssh", func() {
			privateKey, err := ssh.ParsePrivateKey([]byte(bbl.SSHKey()))
			Expect(err).NotTo(HaveOccurred())

			directorAddressURL, err := url.Parse(bbl.DirectorAddress())
			Expect(err).NotTo(HaveOccurred())

			address := fmt.Sprintf("%s:22", directorAddressURL.Hostname())
			_, err = ssh.Dial("tcp", address,
				&ssh.ClientConfig{
					User: "jumpbox",
					Auth: []ssh.AuthMethod{
						ssh.PublicKeys(privateKey),
					},
				})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
