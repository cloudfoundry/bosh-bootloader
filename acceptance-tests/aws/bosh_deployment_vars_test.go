package acceptance_test

import (
	"fmt"
	"time"

	yaml "gopkg.in/yaml.v2"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const ipRegex = `[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`

var _ = Describe("bosh deployment vars", func() {
	var (
		bbl           actors.BBL
		configuration acceptance.Config
	)

	BeforeEach(func() {
		var err error
		configuration, err = acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "bosh-deployment-vars-env")
	})

	AfterEach(func() {
		session := bbl.Destroy()
		Eventually(session, 10*time.Minute).Should(gexec.Exit())
	})

	It("prints the bosh deployment vars for bosh create-env", func() {
		acceptance.SkipUnless("bosh-deployment-vars")
		session := bbl.Up("--name", bbl.PredefinedEnvID(), "--no-director")
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

		stdout := bbl.BOSHDeploymentVars()

		var vars struct {
			InternalCIDR          string   `yaml:"internal_cidr"`
			InternalGateway       string   `yaml:"internal_gw"`
			InternalIP            string   `yaml:"internal_ip"`
			DirectorName          string   `yaml:"director_name"`
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
		Expect(vars.AZ).To(MatchRegexp(`us-.+-\d\w`))
		Expect(vars.SubnetID).To(MatchRegexp("subnet-.+"))
		Expect(vars.AccessKeyID).To(MatchRegexp(".{20}"))
		Expect(vars.SecretAccessKey).To(MatchRegexp(".{40}"))
		Expect(vars.DefaultKeyName).To(Equal(fmt.Sprintf("%s_bosh_vms", bbl.PredefinedEnvID())))
		Expect(vars.DefaultSecurityGroups).To(ContainElement(MatchRegexp("sg-.+")))
		Expect(vars.Region).To(Equal(configuration.AWSRegion))
		Expect(vars.PrivateKey).To(MatchRegexp(`-----BEGIN RSA PRIVATE KEY-----(.*\n)*-----END RSA PRIVATE KEY-----`))
	})
})
