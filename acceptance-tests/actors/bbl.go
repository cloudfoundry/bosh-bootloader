package actors

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/kr/pty"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

type BBL struct {
	stateDirectory    string
	pathToBBL         string
	configuration     acceptance.Config
	envID             string
	useBBLStateBucket bool
}

func NewBBL(stateDirectory string, pathToBBL string, configuration acceptance.Config, envIDSuffix string, useBBLStateBucket bool) BBL {
	envIDPrefix := os.Getenv("BBL_TEST_ENV_ID_PREFIX")
	if envIDPrefix == "" {
		envIDPrefix = "bbl-test"
	}

	return BBL{
		stateDirectory:    stateDirectory,
		pathToBBL:         pathToBBL,
		configuration:     configuration,
		envID:             fmt.Sprintf("%s-%s", envIDPrefix, envIDSuffix),
		useBBLStateBucket: useBBLStateBucket,
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

func (b BBL) VerifySSH(session *gexec.Session, ptmx *os.File) {
	time.Sleep(5 * time.Second)
	fmt.Fprintln(ptmx, "whoami | rev")
	Eventually(session.Out, 10).Should(gbytes.Say("xobpmuj")) // jumpbox in reverse
	fmt.Fprintln(ptmx, "exit 0")
	Eventually(session, 5).Should(gexec.Exit(0))
}

func (b BBL) JumpboxSSH(output io.Writer) *exec.Cmd {
	return b.interactiveExecute([]string{
		"--state-dir", b.stateDirectory,
		"--debug",
		"ssh",
		"--jumpbox",
	}, output, output, b.VerifySSH)
}

func (b BBL) DirectorSSH(output io.Writer) *exec.Cmd {
	return b.interactiveExecute([]string{
		"--state-dir", b.stateDirectory,
		"--debug",
		"ssh",
		"--director",
	}, output, output, b.VerifySSH)
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
	if b.useBBLStateBucket {
		return b.fetchValueFromRemoteBBLState(value)
	}
	return b.fetchValueFromLocalBBLState(value)
}

func (b BBL) fetchValueFromRemoteBBLState(value string) string {
	args := []string{
		"--name", b.envID,
		"--state-bucket", b.configuration.BBLStateBucket,
		value,
	}

	stdout := bytes.NewBuffer([]byte{})
	stderr := bytes.NewBuffer([]byte{})
	b.execute(args, stdout, stderr).Wait(30 * time.Second)
	fmt.Println(stderr)

	return strings.TrimSpace(string(stdout.Bytes()))
}
func (b BBL) fetchValueFromLocalBBLState(value string) string {
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

// partially cribbed from https://github.com/cloudfoundry/vizzini/blob/master/ssh_test.go#L473
func (b BBL) interactiveExecute(args []string, stdout io.Writer, stderr io.Writer, actions func(*gexec.Session, *os.File)) *exec.Cmd {
	cmd := exec.Command(b.pathToBBL, args...)

	ptmx, pts, err := pty.Open()
	Expect(err).NotTo(HaveOccurred())
	defer ptmx.Close()

	cmd.Stdin = pts
	cmd.Stdout = pts
	cmd.Stderr = pts

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setctty: true,
		Setsid:  true,
	}

	session, err := gexec.Start(cmd, stdout, stderr)
	Expect(err).NotTo(HaveOccurred())

	// Close our open reference to pts so that ptmx recieves EOF
	pts.Close()

	actions(session, ptmx)

	return cmd
}
