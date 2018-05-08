package actors

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/kr/pty"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

type BBL struct {
	stateDirectory string
	pathToBBL      string
	configuration  acceptance.Config
	envID          string
}

func NewBBL(stateDirectory string, pathToBBL string, configuration acceptance.Config, envIDSuffix string) BBL {
	envIDPrefix := os.Getenv("BBL_TEST_ENV_ID_PREFIX")
	if envIDPrefix == "" {
		envIDPrefix = "bbl-test"
	}

	return BBL{
		stateDirectory: stateDirectory,
		pathToBBL:      pathToBBL,
		configuration:  configuration,
		envID:          fmt.Sprintf("%s-%s", envIDPrefix, envIDSuffix),
	}
}

func (b BBL) PredefinedEnvID() string {
	return b.envID
}

func (b BBL) Up(additionalArgs ...string) *gexec.Session {
	args := []string{
		"--state-dir", b.stateDirectory,
		"--debug",
		"up",
	}

	args = append(args, additionalArgs...)

	return b.execute(args, os.Stdout, os.Stderr)
}

func (b BBL) Plan(additionalArgs ...string) *gexec.Session {
	args := []string{
		"--state-dir", b.stateDirectory,
		"--debug",
		"plan",
	}

	args = append(args, additionalArgs...)

	return b.execute(args, os.Stdout, os.Stderr)
}

func (b BBL) Rotate() *gexec.Session {
	return b.execute([]string{
		"--state-dir", b.stateDirectory,
		"--debug",
		"rotate",
	}, os.Stdout, os.Stderr)
}

func (b BBL) VerifySSH(sshFunc func() (*exec.Cmd, *os.File)) {
	cmd, session := sshFunc()
	defer session.Close()

	time.Sleep(5 * time.Second)
	fmt.Fprintln(session, "whoami")
	fmt.Fprintln(session, "exit 0")
	time.Sleep(5 * time.Second)
	output, err := ioutil.ReadAll(session)
	Expect(err).NotTo(HaveOccurred())
	Expect(string(output)).To(ContainSubstring("jumpbox"))

	Eventually(cmd.Wait).Should(Succeed(), fmt.Sprintf("output was:\n\n%s", output))
}

func (b BBL) JumpboxSSH() (*exec.Cmd, *os.File) {
	return b.interactiveExecute([]string{
		"--state-dir", b.stateDirectory,
		"--debug",
		"ssh",
		"--jumpbox",
	}, os.Stdout, os.Stderr)
}

func (b BBL) DirectorSSH() (*exec.Cmd, *os.File) {
	return b.interactiveExecute([]string{
		"--state-dir", b.stateDirectory,
		"--debug",
		"ssh",
		"--director",
	}, os.Stdout, os.Stderr)
}

func (b BBL) Destroy() *gexec.Session {
	return b.execute([]string{
		"--state-dir", b.stateDirectory,
		"--debug",
		"destroy",
		"--no-confirm",
	}, os.Stdout, os.Stderr)
}

func (b BBL) Down() *gexec.Session {
	return b.execute([]string{
		"--state-dir", b.stateDirectory,
		"--debug",
		"down",
		"--no-confirm",
	}, os.Stdout, os.Stderr)
}

func (b BBL) CleanupLeftovers(filter string) *gexec.Session {
	return b.execute([]string{
		"--state-dir", b.stateDirectory,
		"cleanup-leftovers",
		"--filter", filter,
		"--no-confirm",
	}, os.Stdout, os.Stderr)
}

func (b BBL) Lbs() string {
	return b.fetchValue("lbs")
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

func (b BBL) DirectorCACert() string {
	return b.fetchValue("director-ca-cert")
}

func (b BBL) JumpboxAddress() string {
	return b.fetchValue("jumpbox-address")
}

func (b BBL) SSHKey() string {
	return b.fetchValue("ssh-key")
}

func (b BBL) DirectorSSHKey() string {
	return b.fetchValue("director-ssh-key")
}

func (b BBL) EnvID() string {
	return b.fetchValue("env-id")
}

func (b BBL) PrintEnv() string {
	return b.fetchValue("print-env")
}

func (b BBL) LatestError() string {
	return b.fetchValue("latest-error")
}

func (b BBL) SaveDirectorCA() string {
	stdout := bytes.NewBuffer([]byte{})
	session := b.execute([]string{
		"--state-dir", b.stateDirectory,
		"director-ca-cert",
	}, stdout, os.Stderr)
	Eventually(session, 10*time.Minute).Should(gexec.Exit(0))

	file, err := ioutil.TempFile("", "")
	defer file.Close()
	Expect(err).NotTo(HaveOccurred())

	file.Write(stdout.Bytes())

	return file.Name()
}

func (b BBL) ExportBoshAllProxy() string {
	lines := strings.Split(b.PrintEnv(), "\n")
	value := getExport("BOSH_ALL_PROXY", lines)
	os.Setenv("BOSH_ALL_PROXY", value)
	return value
}

func (b BBL) StartSSHTunnel() *gexec.Session {
	printEnvLines := strings.Split(b.PrintEnv(), "\n")
	os.Setenv("BOSH_ALL_PROXY", getExport("BOSH_ALL_PROXY", printEnvLines))

	var sshArgs []string
	for i := 0; i < len(printEnvLines); i++ {
		if strings.HasPrefix(printEnvLines[i], "ssh ") {
			sshCmd := strings.TrimPrefix(printEnvLines[i], "ssh ")
			sshCmd = strings.Replace(sshCmd, "$JUMPBOX_PRIVATE_KEY", getExport("JUMPBOX_PRIVATE_KEY", printEnvLines), -1)
			sshCmd = strings.Replace(sshCmd, "-f ", "", -1)
			sshArgs = strings.Split(sshCmd, " ")
		}
	}

	cmd := exec.Command("ssh", sshArgs...)
	sshSession, err := gexec.Start(cmd, ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	return sshSession
}

func getExport(keyName string, lines []string) string {
	for _, line := range lines {
		if strings.HasPrefix(line, fmt.Sprintf("export %s", keyName)) {
			parts := strings.Split(line, " ")
			keyValue := parts[1]
			keyValueParts := strings.Split(keyValue, "=")
			return strings.Join(keyValueParts[1:], "=")
		}
	}
	return ""
}

func (b BBL) fetchValue(value string) string {
	args := []string{
		"--state-dir", b.stateDirectory,
		value,
	}

	stdout := bytes.NewBuffer([]byte{})
	stderr := bytes.NewBuffer([]byte{})
	b.execute(args, stdout, stderr).Wait(30 * time.Second)

	return strings.TrimSpace(string(stdout.Bytes()))
}

func (b BBL) execute(args []string, stdout io.Writer, stderr io.Writer) *gexec.Session {
	cmd := exec.Command(b.pathToBBL, args...)
	session, err := gexec.Start(cmd, stdout, stderr)
	Expect(err).NotTo(HaveOccurred())

	return session
}

func (b BBL) interactiveExecute(args []string, stdout io.Writer, stderr io.Writer) (*exec.Cmd, *os.File) {
	cmd := exec.Command(b.pathToBBL, args...)
	f, err := pty.Start(cmd)
	Expect(err).NotTo(HaveOccurred())

	return cmd, f
}
