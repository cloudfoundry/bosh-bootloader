package acceptance_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"
	proxy "github.com/cloudfoundry/socks5-proxy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("up after plan", func() {
	var (
		bbl        actors.BBL
		stateDir   string
		jumpboxURL string
	)

	BeforeEach(func() {
		acceptance.SkipUnless("up-after-plan")

		configuration, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		stateDir = configuration.StateFileDir

		bbl = actors.NewBBL(stateDir, pathToBBL, configuration, "up-after-plan-env")
	})

	AfterEach(func() {
		deleteDirectorPath := filepath.Join(stateDir, "delete-director.sh")
		deleteJumpboxPath := filepath.Join(stateDir, "delete-jumpbox.sh")
		noOpScript := []byte("#!/bin/bash\n")

		err := ioutil.WriteFile(deleteDirectorPath, noOpScript, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(deleteJumpboxPath, noOpScript, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		session := bbl.Down()
		Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
	})

	It("preserves files modified after plan", func() {
		createEnvOutputPath := filepath.Join(stateDir, "create-env-output")
		By("running bbl plan", func() {
			session := bbl.Plan("--name", bbl.PredefinedEnvID())
			Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
		})

		By("starting an SSH server as a double for the jumpbox", func() {
			httpServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			}))
			httpServerHostPort := strings.Split(httpServer.URL, "http://")[1]

			err := exec.Command("bosh",
				"int", filepath.Join(stateDir, "jumpbox-deployment", "jumpbox.yml"),
				"--vars-store", filepath.Join(stateDir, "vars", "jumpbox-variables.yml"),
			).Run()
			Expect(err).NotTo(HaveOccurred())

			vars, err := ioutil.ReadFile(filepath.Join(stateDir, "vars", "jumpbox-variables.yml"))
			Expect(err).NotTo(HaveOccurred())
			key := getJumpboxPrivateKey(string(vars))
			jumpboxURL = proxy.StartTestSSHServer(httpServerHostPort, key)
		})

		By("modifying the plan", func() {
			createDirectorPath := filepath.Join(stateDir, "create-director.sh")
			newCreateDirector := []byte(fmt.Sprintf("#!/bin/bash\necho 'director' >> %s\n", createEnvOutputPath))
			err := ioutil.WriteFile(createDirectorPath, newCreateDirector, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			createJumpboxPath := filepath.Join(stateDir, "create-jumpbox.sh")
			newCreateJumpbox := []byte(fmt.Sprintf("#!/bin/bash\necho 'jumpbox' >> %s\n", createEnvOutputPath))
			err = ioutil.WriteFile(createJumpboxPath, newCreateJumpbox, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			terraformTemplatePath := filepath.Join(stateDir, "terraform", "template.tf")
			newTerraformTemplate := []byte(fmt.Sprintf(`
output "jumpbox_url" {
	value = "%s"
}
`, jumpboxURL))
			err = ioutil.WriteFile(terraformTemplatePath, newTerraformTemplate, os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		By("running bbl up", func() {
			time.Sleep(5 * time.Second)
			session := bbl.Up()
			// Don't check the exit code of up because upload cloud config fails.
			// We don't yet have a way to inject different behavior for that step.
			Eventually(session, 40*time.Minute).Should(gexec.Exit())
		})

		By("verifying that modified scripts were run", func() {
			createEnvOutput, err := ioutil.ReadFile(createEnvOutputPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(createEnvOutput)).To(Equal("jumpbox\ndirector\n"))
		})
	})
})

func getJumpboxPrivateKey(v string) string {
	var vars struct {
		JumpboxSSH struct {
			PrivateKey string `yaml:"private_key"`
		} `yaml:"jumpbox_ssh"`
	}

	err := yaml.Unmarshal([]byte(v), &vars)
	Expect(err).NotTo(HaveOccurred())

	return vars.JumpboxSSH.PrivateKey
}
