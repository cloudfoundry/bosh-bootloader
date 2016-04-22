package actors

import (
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test"
)

type BBL struct {
	stateDirectory string
	pathToBBL      string
	configuration  integration.Config
}

func NewBBL(stateDirectory string, pathToBBL string, configuration integration.Config) BBL {
	return BBL{
		stateDirectory: stateDirectory,
		pathToBBL:      pathToBBL,
		configuration:  configuration,
	}
}

func (b BBL) Up(loadBalancerType string) {
	session := b.execute([]string{
		"--aws-access-key-id", b.configuration.AWSAccessKeyID,
		"--aws-secret-access-key", b.configuration.AWSSecretAccessKey,
		"--aws-region", b.configuration.AWSRegion,
		"--state-dir", b.stateDirectory,
		"unsupported-deploy-bosh-on-aws-for-concourse",
		"--lb-type", loadBalancerType,
	})
	Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
}

func (b BBL) Destroy() {
	session := b.execute([]string{
		"--state-dir", b.stateDirectory,
		"destroy",
		"--no-confirm",
	})
	Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
}

func (b BBL) DirectorUsername() string {
	return b.fetchValue("director-username")
}

func (b BBL) DirectorPassword() string {
	return b.fetchValue("director-password")
}

func (b BBL) DirectorAddress() string {
	return b.fetchValue("director-address")
}

func (b BBL) fetchValue(value string) string {
	session := b.execute([]string{
		"--state-dir", b.stateDirectory,
		value,
	})
	return strings.TrimSpace(string(session.Wait().Buffer().Contents()))
}

func (b BBL) execute(args []string) *gexec.Session {
	cmd := exec.Command(b.pathToBBL, args...)
	session, err := gexec.Start(cmd, os.Stdout, os.Stderr)
	Expect(err).NotTo(HaveOccurred())

	return session
}
