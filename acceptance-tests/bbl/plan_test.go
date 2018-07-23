package acceptance_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("plan", func() {
	var (
		bbl      actors.BBL
		stateDir string
		iaas     string
	)

	BeforeEach(func() {
		acceptance.SkipUnless("plan")

		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		iaas = configuration.IAAS
		stateDir = configuration.StateFileDir

		bbl = actors.NewBBL(stateDir, pathToBBL, configuration, "plan-env", false)
	})

	It("sets up the bbl state directory", func() {
		session := bbl.Plan("--name", bbl.PredefinedEnvID())
		Eventually(session, bblPlanTimeout).Should(gexec.Exit(0))

		expectedArtifacts := []string{
			filepath.Join(stateDir, "create-jumpbox.sh"),
			filepath.Join(stateDir, "create-director.sh"),
			filepath.Join(stateDir, "delete-jumpbox.sh"),
			filepath.Join(stateDir, "delete-director.sh"),
			filepath.Join(stateDir, "cloud-config", "cloud-config.yml"),
			filepath.Join(stateDir, "cloud-config", "ops.yml"),
			filepath.Join(stateDir, "bosh-deployment", "bosh.yml"),
			filepath.Join(stateDir, "bosh-deployment", iaas, "cpi.yml"),
			filepath.Join(stateDir, "bosh-deployment", "credhub.yml"),
			filepath.Join(stateDir, "bosh-deployment", "jumpbox-user.yml"),
			filepath.Join(stateDir, "bosh-deployment", "uaa.yml"),
			filepath.Join(stateDir, "jumpbox-deployment", "jumpbox.yml"),
			filepath.Join(stateDir, "jumpbox-deployment", iaas, "cpi.yml"),
			filepath.Join(stateDir, "terraform", "bbl-template.tf"),
		}

		By("verifying that artifacts are created in state dir", func() {
			for _, f := range expectedArtifacts {
				fileinfo, err := os.Stat(f)
				Expect(err).NotTo(HaveOccurred())
				if strings.HasSuffix(f, ".sh") {
					Expect(fileinfo.Mode().String()).To(Equal("-rwxr-x---"))
				} else {
					Expect(fileinfo.Mode().String()).To(Equal("-rwxr-----"))
				}
			}
		})

		By("modifying artifacts", func() {
			for _, f := range expectedArtifacts {
				err := ioutil.WriteFile(f, []byte("modified after plan"), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		var patchFileHandle *os.File
		patchFile := filepath.Join(stateDir, "terraform", "patch.tf")
		By("adding a patch file", func() {
			var err error
			patchFileHandle, err = os.Create(patchFile)
			Expect(err).NotTo(HaveOccurred())
		})
		defer patchFileHandle.Close()

		By("rerunning bbl plan", func() {
			session = bbl.Plan("--name", bbl.PredefinedEnvID())
			Eventually(session, bblPlanTimeout).Should(gexec.Exit(0))
		})

		By("verifying that modified artifacts were overwritten", func() {
			for _, f := range expectedArtifacts {
				contents, err := ioutil.ReadFile(f)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(contents)).NotTo(ContainSubstring("modified after plan"))
			}
		})

		By("verifying that we warn users about patch files", func() {
			Expect(session.Out).To(gbytes.Say("\nyou've supplied the following files to bbl:\n"))
			Expect(session.Out).To(gbytes.Say("\tterraform/patch.tf\n"))
			Expect(session.Out).To(gbytes.Say("\nthey will be used by \"bbl up\".\n"))
		})

		By("running bbl down", func() {
			session = bbl.Down()
			Eventually(session, bblDownTimeout).Should(gexec.Exit(0))
		})

		By("verifying that artifacts were deleted", func() {
			for _, f := range expectedArtifacts {
				_, err := os.Stat(f)
				Expect(err).To(HaveOccurred())
			}
		})

		By("verifying that patch files were not deleted", func() {
			_, err := os.Stat(patchFile)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
