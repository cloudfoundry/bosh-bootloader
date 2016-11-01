package main_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/cloudfoundry/bosh-bootloader/bbl/awsbackend"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/rosenhouse/awsfaker"
)

const (
	LinuxReleaseURL = "https://github.com/cloudfoundry/bosh-bootloader/releases/download/v1.0.0/bbl-v1.0.0_linux_x86-64"
	MacReleaseURL   = "https://github.com/cloudfoundry/bosh-bootloader/releases/download/v1.0.0/bbl-v1.0.0_osx"
)

var _ = Describe("upgrade from 1.0.0", func() {
	var (
		tmpDir      string
		envID       string
		pathToBBLv1 string

		fakeBOSH       *fakeBOSHDirector
		fakeBOSHServer *httptest.Server
		fakeAWS        *awsbackend.Backend
		fakeAWSServer  *httptest.Server
	)

	BeforeEach(func() {
		var err error

		tmpDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		fakeBOSH = &fakeBOSHDirector{}
		fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			fakeBOSH.ServeHTTP(responseWriter, request)
		}))

		fakeAWS = awsbackend.New(fakeBOSHServer.URL)
		fakeAWSServer = httptest.NewServer(awsfaker.New(fakeAWS))
	})

	bblUpV1 := func(bbl string) (*gexec.Session, error) {
		args := []string{
			fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
			"--state-dir", tmpDir,
			"up",
			"--aws-access-key-id", "some-access-key",
			"--aws-secret-access-key", "some-access-secret",
			"--aws-region", "some-region",
		}
		session, err := gexec.Start(exec.Command(bbl, args...), GinkgoWriter, GinkgoWriter)
		return session, err
	}

	bblUpDev := func(bbl string) (*gexec.Session, error) {
		args := []string{
			fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
			"--state-dir", tmpDir,
			"up",
			"--iaas", "aws",
			"--aws-access-key-id", "some-access-key",
			"--aws-secret-access-key", "some-access-secret",
			"--aws-region", "some-region",
		}
		session, err := gexec.Start(exec.Command(bbl, args...), GinkgoWriter, GinkgoWriter)
		return session, err
	}

	downloadBBLV1 := func() error {
		var releaseURL string
		if runtime.GOOS == "linux" {
			releaseURL = LinuxReleaseURL
		} else {
			releaseURL = MacReleaseURL
		}

		resp, err := http.Get(releaseURL)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			// FIXME add more info to the error
			return errors.New("failed to download bbl v1")
		}

		bblBinary, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		pathToBBLv1 = filepath.Join(tmpDir, "bbl")
		err = ioutil.WriteFile(pathToBBLv1, bblBinary, os.ModePerm)
		if err != nil {
			return err
		}

		return nil
	}

	It("maintains env id", func() {
		By("downloading bbl v1.0.0", func() {
			err := downloadBBLV1()
			Expect(err).NotTo(HaveOccurred())
		})

		By("bbl-ing up with v1.0.0", func() {
			session, err := bblUpV1(pathToBBLv1)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		})

		By("retrieving the env-id from state file", func() {
			args := []string{"--state-dir", tmpDir, "env-id"}
			session, err := gexec.Start(exec.Command(pathToBBLv1, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 3*time.Second).Should(gexec.Exit(0))

			stdout := session.Out.Contents()
			envID = string(stdout)
		})

		By("bbl-ing up with dev version", func() {
			session, err := bblUpDev(pathToBBL)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 10*time.Second).Should(gexec.Exit(0))
		})

		By("asserting that the env-id has not changed", func() {
			args := []string{"--state-dir", tmpDir, "env-id"}
			session, err := gexec.Start(exec.Command(pathToBBL, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session, 3*time.Second).Should(gexec.Exit(0))

			stdout := session.Out.Contents()
			Expect(string(stdout)).To(Equal(envID))
		})
	})
})
