package actors

import (
	"bytes"
	"io"
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
	args := []string{
		"--aws-access-key-id", b.configuration.AWSAccessKeyID,
		"--aws-secret-access-key", b.configuration.AWSSecretAccessKey,
		"--aws-region", b.configuration.AWSRegion,
		"--state-dir", b.stateDirectory,
		"unsupported-deploy-bosh-on-aws-for-concourse",
		"--lb-type", loadBalancerType,
	}

	if loadBalancerType == "cf" || loadBalancerType == "concourse" {
		args = append(args, []string{
			"--cert", "bbl-certs/bbl.crt",
			"--key", "bbl-certs/bbl.key",
		}...)
	}

	session := b.execute(args, os.Stdout, os.Stderr)
	Eventually(session, 40*time.Minute).Should(gexec.Exit(0))
}

func (b BBL) Destroy() {
	session := b.execute([]string{
		"--state-dir", b.stateDirectory,
		"destroy",
		"--no-confirm",
	}, os.Stdout, os.Stderr)
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

func (b BBL) CreateLB(loadBalancerType string) {
	args := []string{
		"--aws-access-key-id", b.configuration.AWSAccessKeyID,
		"--aws-secret-access-key", b.configuration.AWSSecretAccessKey,
		"--aws-region", b.configuration.AWSRegion,
		"--state-dir", b.stateDirectory,
		"unsupported-create-lbs",
		"--type", loadBalancerType,
		"--cert", "bbl-certs/bbl-intermediate.crt",
		"--key", "bbl-certs/bbl-intermediate.key",
		"--chain", "bbl-certs/bbl.crt",
	}

	session := b.execute(args, os.Stdout, os.Stderr)
	Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
}

func (b BBL) LBs() *gexec.Session {
	args := []string{
		"--aws-access-key-id", b.configuration.AWSAccessKeyID,
		"--aws-secret-access-key", b.configuration.AWSSecretAccessKey,
		"--aws-region", b.configuration.AWSRegion,
		"--state-dir", b.stateDirectory,
		"unsupported-lbs",
	}

	session := b.execute(args, os.Stdout, os.Stderr)
	Eventually(session, 10*time.Minute).Should(gexec.Exit(0))

	return session
}

func (b BBL) UpdateLB(certPath, keyPath string) {
	args := []string{
		"--aws-access-key-id", b.configuration.AWSAccessKeyID,
		"--aws-secret-access-key", b.configuration.AWSSecretAccessKey,
		"--aws-region", b.configuration.AWSRegion,
		"--state-dir", b.stateDirectory,
		"unsupported-update-lbs",
		"--cert", certPath,
		"--key", keyPath,
	}

	session := b.execute(args, os.Stdout, os.Stderr)
	Eventually(session, 10*time.Minute).Should(gexec.Exit(0))
}

func (b BBL) fetchValue(value string) string {
	args := []string{
		"--state-dir", b.stateDirectory,
		value,
	}

	stdout := bytes.NewBuffer([]byte{})
	stderr := bytes.NewBuffer([]byte{})
	b.execute(args, stdout, stderr).Wait()

	return strings.TrimSpace(string(stdout.Bytes()))
}

func (b BBL) execute(args []string, stdout io.Writer, stderr io.Writer) *gexec.Session {
	cmd := exec.Command(b.pathToBBL, args...)
	session, err := gexec.Start(cmd, stdout, stderr)
	Expect(err).NotTo(HaveOccurred())

	return session
}
