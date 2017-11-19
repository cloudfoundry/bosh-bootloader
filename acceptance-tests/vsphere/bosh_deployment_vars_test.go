package acceptance_test

import (
	"fmt"
	"time"

	yaml "gopkg.in/yaml.v2"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"
	"github.com/cloudfoundry/bosh-bootloader/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const ipRegex = `[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`

var _ = Describe("bosh deployment vars", func() {
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

		session := bbl.Up("--name", bbl.PredefinedEnvID(), "--no-director")
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
	})

	AfterEach(func() {
		session := bbl.Destroy()
		Eventually(session, 10*time.Minute).Should(gexec.Exit())
	})

	It("prints the bosh deployment vars for bosh create-env", func() {
		stdout := bbl.BOSHDeploymentVars()

		var vars struct {
			InternalCIDR     string `yaml:"internal_cidr"`
			InternalGateway  string `yaml:"internal_gw"`
			InternalIP       string `yaml:"internal_ip"`
			ExternalIP       string `yaml:"external_ip"`
			DirectorName     string `yaml:"director_name"`
			NetworkName      string `yaml:"network_name"`
			VCenterUser      string `yaml:"vcenter_user"`
			VCenterPassword  string `yaml:"vcenter_password"`
			VCenterDC        string `yaml:"vcenter_dc"`
			VCenterCluster   string `yaml:"vcenter_cluster"`
			VCenterRP        string `yaml:"vcenter_rp"`
			VCenterDS        string `yaml:"vcenter_ds"`
			VCenterDisks     string `yaml:"vcenter_disks"`
			VCenterVMs       string `yaml:"vcenter_vms"`
			VCenterTemplates string `yaml:"vcenter_templates"`
		}

		yaml.Unmarshal([]byte(stdout), &vars)
		fmt.Println(stdout)
		parsedCidr, err := bosh.ParseCIDRBlock(configuration.VSphereSubnet)
		Expect(err).NotTo(HaveOccurred())

		Expect(vars.InternalCIDR).To(Equal(configuration.VSphereSubnet))
		Expect(vars.InternalGateway).To(Equal(parsedCidr.GetNthIP(1).String()))
		Expect(vars.InternalIP).To(Equal(parsedCidr.GetNthIP(6).String()))
		Expect(vars.DirectorName).To(Equal(fmt.Sprintf("bosh-%s", bbl.PredefinedEnvID())))

		Expect(vars.VCenterUser).To(Equal(configuration.VSphereVCenterUser))
		Expect(vars.VCenterPassword).To(Equal(configuration.VSphereVCenterPassword))
		Expect(vars.VCenterDC).To(Equal(configuration.VSphereVCenterDC))
		Expect(vars.VCenterCluster).To(Equal(configuration.VSphereVCenterCluster))
		Expect(vars.VCenterRP).To(Equal(configuration.VSphereVCenterRP))
		Expect(vars.NetworkName).To(Equal(configuration.VSphereNetwork))
		Expect(vars.VCenterDS).To(Equal(configuration.VSphereVCenterDS))
		Expect(vars.VCenterDisks).To(Equal(configuration.VSphereVCenterDisks))
		Expect(vars.VCenterVMs).To(Equal(configuration.VSphereVCenterVMs))
		Expect(vars.VCenterTemplates).To(Equal(configuration.VSphereVCenterTemplates))
	})
})
