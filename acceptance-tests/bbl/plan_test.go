package acceptance_test

import (
	"os"
	"path/filepath"
	"time"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("plan", func() {
	var (
		bbl      actors.BBL
		stateDir string
	)

	BeforeEach(func() {
		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		stateDir = configuration.StateFileDir

		bbl = actors.NewBBL(stateDir, pathToBBL, configuration, "plan-env")
	})

	AfterEach(func() {
		session := bbl.Down()
		Eventually(session, 10*time.Minute).Should(gexec.Exit())
	})

	It("sets up the bbl state directory", func() {
		session := bbl.Plan("--name", bbl.PredefinedEnvID())
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

		By("verifying that artifacts are created in state dir", func() {
			checkExists := func(dir string, filenames []string) {
				for _, f := range filenames {
					_, err := os.Stat(filepath.Join(dir, f))
					Expect(err).NotTo(HaveOccurred())
				}
			}

			checkExists(stateDir, []string{"bbl-state.json"})
			checkExists(stateDir, []string{"create-jumpbox.sh"})
			checkExists(stateDir, []string{"create-director.sh"})
			checkExists(stateDir, []string{"delete-jumpbox.sh"})
			checkExists(stateDir, []string{"delete-director.sh"})
			checkExists(filepath.Join(stateDir, ".bbl", "cloudconfig"), []string{
				"cloud-config.yml",
				"ops.yml",
			})
			checkExists(filepath.Join(stateDir, ".bbl"), []string{
				"previous-user-ops-file.yml",
			})
			checkExists(filepath.Join(stateDir, "bosh-deployment"), []string{
				"bosh.yml",
				"cpi.yml",
				"credhub.yml",
				"jumpbox-user.yml",
				"uaa.yml",
			})
			checkExists(filepath.Join(stateDir, "jumpbox-deployment"), []string{
				"cpi.yml",
				"jumpbox.yml",
			})
			checkExists(filepath.Join(stateDir, "terraform"), []string{
				"template.tf",
			})
			checkExists(filepath.Join(stateDir, "vars"), []string{
				// "bosh-state.json",
				"director-variables.yml",
				// "jumpbox-state.json",
				"jumpbox-variables.yml",
				// "terraform.tfstate",
				"user-ops-file.yml",
			})
		})
	})
})
