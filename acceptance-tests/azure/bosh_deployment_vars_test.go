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

var _ = FDescribe("bosh deployment vars", func() {
	var (
		bbl           actors.BBL
		state         acceptance.State
		configuration acceptance.Config
	)

	BeforeEach(func() {
		var err error
		configuration, err = acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "bosh-deployment-vars-env")
		state = acceptance.NewState(configuration.StateFileDir)

		session := bbl.Up("azure", []string{"--name", bbl.PredefinedEnvID(), "--no-director"})
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
	})

	AfterEach(func() {
		session := bbl.Destroy()
		Eventually(session, 10*time.Minute).Should(gexec.Exit())
	})

	It("prints the bosh deployment vars for bosh create-env", func() {
		stdout := bbl.BOSHDeploymentVars()

		var vars struct {
			InternalCIDR         string `yaml:"internal_cidr"`
			InternalGateway      string `yaml:"internal_gw"`
			InternalIP           string `yaml:"internal_ip"`
			DirectorName         string `yaml:"director_name"`
			VNetName             string `yaml:"vnet_name"`
			SubnetName           string `yaml:"subnet_name"`
			Subscription         string `yaml:"subscription_id"`
			TenantID             string `yaml:"tenant_id"`
			ClientID             string `yaml:"client_id"`
			ClientSecret         string `yaml:"client_secret"`
			ResourceGroupName    string `yaml:"resource_group_name"`
			StoreageAccountName  string `yaml:"storage_account_name"`
			DefaultSecurityGroup string `yaml:"default_security_group"`
		}

		yaml.Unmarshal([]byte(stdout), &vars)
		Expect(vars.InternalCIDR).To(Equal("10.0.0.0/24"))
		Expect(vars.InternalGateway).To(Equal("10.0.0.1"))
		Expect(vars.InternalIP).To(Equal("10.0.0.6"))
		Expect(vars.DirectorName).To(Equal(fmt.Sprintf("bosh-%s", bbl.PredefinedEnvID())))
	})
})
