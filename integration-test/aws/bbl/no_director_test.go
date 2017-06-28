package integration_test

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	integration "github.com/cloudfoundry/bosh-bootloader/integration-test"
	"github.com/cloudfoundry/bosh-bootloader/integration-test/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const ipRegex = `[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`

var _ = Describe("no director test", func() {
	var (
		bbl           actors.BBL
		aws           actors.AWS
		state         integration.State
		configuration integration.Config
	)

	BeforeEach(func() {
		var err error
		configuration, err = integration.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "no-director-env")
		aws = actors.NewAWS(configuration)
		state = integration.NewState(configuration.StateFileDir)
	})

	AfterEach(func() {
		if !CurrentGinkgoTestDescription().Failed {
			bbl.Destroy()
		}
	})

	It("successfully standups up a no director infrastructure", func() {
		By("calling bbl up with the no-director flag", func() {
			bbl.Up(actors.AWSIAAS, []string{"--name", bbl.PredefinedEnvID(), "--no-director"})
		})

		By("checking that an instance exists", func() {
			instances := aws.Instances(fmt.Sprintf("%s-vpc", bbl.PredefinedEnvID()))
			Expect(instances).To(HaveLen(1))
			Expect(instances).To(Equal([]string{fmt.Sprintf("%s-nat", bbl.PredefinedEnvID())}))

			tags := aws.GetEC2InstanceTags(fmt.Sprintf("%s-nat", bbl.PredefinedEnvID()))
			Expect(tags["EnvID"]).To(Equal(bbl.PredefinedEnvID()))
		})

		By("checking that director details are not printed", func() {
			directorUsername := bbl.DirectorUsername()
			Expect(directorUsername).To(Equal(""))
		})

		By("checking that bosh-deployment-vars prints a bosh create-env compatible vars-file", func() {
			stdout := bbl.BOSHDeploymentVars()

			var vars struct {
				InternalCIDR          string   `yaml:"internal_cidr"`
				InternalGateway       string   `yaml:"internal_gw"`
				InternalIP            string   `yaml:"internal_ip"`
				DirectorName          string   `yaml:"director_name"`
				ExternalIP            string   `yaml:"external_ip"`
				AZ                    string   `yaml:"az"`
				SubnetID              string   `yaml:"subnet_id"`
				AccessKeyID           string   `yaml:"access_key_id"`
				SecretAccessKey       string   `yaml:"secret_access_key"`
				DefaultKeyName        string   `yaml:"default_key_name"`
				DefaultSecurityGroups []string `yaml:"default_security_groups"`
				Region                string   `yaml:"region"`
				PrivateKey            string   `yaml:"private_key"`
			}

			yaml.Unmarshal([]byte(stdout), &vars)
			Expect(vars.InternalCIDR).To(Equal("10.0.0.0/24"))
			Expect(vars.InternalGateway).To(Equal("10.0.0.1"))
			Expect(vars.InternalIP).To(Equal("10.0.0.6"))
			Expect(vars.DirectorName).To(Equal(fmt.Sprintf("bosh-%s", bbl.PredefinedEnvID())))
			Expect(vars.ExternalIP).To(MatchRegexp(ipRegex))
			Expect(vars.AZ).To(MatchRegexp(`us-.+-\d\w`))
			Expect(vars.SubnetID).To(MatchRegexp("subnet-.+"))
			Expect(vars.AccessKeyID).To(MatchRegexp(".{20}"))
			Expect(vars.SecretAccessKey).To(MatchRegexp(".{40}"))
			Expect(vars.DefaultKeyName).To(Equal(fmt.Sprintf("keypair-%s", bbl.PredefinedEnvID())))
			Expect(vars.DefaultSecurityGroups).To(ContainElement(MatchRegexp("sg-.+")))
			Expect(vars.Region).To(Equal(configuration.AWSRegion))
			Expect(vars.PrivateKey).To(MatchRegexp(`-----BEGIN RSA PRIVATE KEY-----(.*\n)*-----END RSA PRIVATE KEY-----`))
		})

		By("checking if bbl print-env prints the external ip", func() {
			stdout := bbl.PrintEnv()

			Expect(stdout).To(ContainSubstring("export BOSH_ENVIRONMENT="))
			Expect(stdout).NotTo(ContainSubstring("export BOSH_CLIENT="))
			Expect(stdout).NotTo(ContainSubstring("export BOSH_CLIENT_SECRET="))
			Expect(stdout).NotTo(ContainSubstring("export BOSH_CA_CERT="))
		})
	})
})
