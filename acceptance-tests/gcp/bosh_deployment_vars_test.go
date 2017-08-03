package acceptance_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const ipRegex = `[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}`

var _ = Describe("bosh deployment vars", func() {
	var (
		bbl                  actors.BBL
		state                acceptance.State
		configuration        acceptance.Config
		gcpServiceAccountKey map[string]interface{}
	)

	BeforeEach(func() {
		var err error
		configuration, err = acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(configuration.StateFileDir, pathToBBL, configuration, "bosh-deployment-vars-env")
		state = acceptance.NewState(configuration.StateFileDir)

		bbl.Up("gcp", []string{"--name", bbl.PredefinedEnvID(), "--no-director"})

		gcpServiceAccountKeyContents, err := ioutil.ReadFile(configuration.GCPServiceAccountKey)
		if err != nil {
			gcpServiceAccountKeyContents = []byte(configuration.GCPServiceAccountKey)
		}

		err = json.Unmarshal(gcpServiceAccountKeyContents, &gcpServiceAccountKey)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		bbl.Destroy()
	})

	It("prints the bosh deployment vars for bosh create-env", func() {
		stdout := bbl.BOSHDeploymentVars()

		var vars struct {
			InternalCIDR       string   `yaml:"internal_cidr"`
			InternalGateway    string   `yaml:"internal_gw"`
			InternalIP         string   `yaml:"internal_ip"`
			DirectorName       string   `yaml:"director_name"`
			ExternalIP         string   `yaml:"external_ip"`
			Zone               string   `yaml:"zone"`
			Network            string   `yaml:"network"`
			Subnetwork         string   `yaml:"subnetwork"`
			Tags               []string `yaml:"tags"`
			ProjectID          string   `yaml:"project_id"`
			GCPCredentialsJSON string   `yaml:"gcp_credentials_json"`
		}

		err := yaml.Unmarshal([]byte(stdout), &vars)
		Expect(err).NotTo(HaveOccurred())

		var returnedAccountKey map[string]interface{}
		err = json.Unmarshal([]byte(vars.GCPCredentialsJSON), &returnedAccountKey)
		Expect(err).NotTo(HaveOccurred())

		Expect(vars.InternalCIDR).To(Equal("10.0.0.0/24"))
		Expect(vars.InternalGateway).To(Equal("10.0.0.1"))
		Expect(vars.InternalIP).To(Equal("10.0.0.6"))
		Expect(vars.DirectorName).To(Equal(fmt.Sprintf("bosh-%s", bbl.PredefinedEnvID())))
		Expect(vars.ExternalIP).To(MatchRegexp(ipRegex))
		Expect(vars.Zone).To(MatchRegexp(`us-.+\d-\w`))
		Expect(vars.Network).To(Equal(fmt.Sprintf("%s-network", bbl.PredefinedEnvID())))
		Expect(vars.Subnetwork).To(Equal(fmt.Sprintf("%s-subnet", bbl.PredefinedEnvID())))
		Expect(vars.Tags).To(Equal([]string{
			fmt.Sprintf("%s-bosh-open", bbl.PredefinedEnvID()),
			fmt.Sprintf("%s-bosh-director", bbl.PredefinedEnvID()),
		}))
		Expect(vars.ProjectID).To(Equal(configuration.GCPProjectID))
		Expect(returnedAccountKey).To(Equal(gcpServiceAccountKey))
	})
})
