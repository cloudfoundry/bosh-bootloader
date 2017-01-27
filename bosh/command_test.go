package bosh_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cloudfoundry/bosh-bootloader/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Cmd", func() {
	var (
		stdout *bytes.Buffer
		stderr *bytes.Buffer

		cmd          bosh.Cmd
		cmdWithDebug bosh.Cmd

		fakeBOSHBackendServer *httptest.Server
		pathToFakeBOSH        string
		pathToBOSH            string

		fastFailBOSH      bool
		fastFailBOSHMutex sync.Mutex

		boshArgs      []string
		boshArgsMutex sync.Mutex
	)

	var setFastFailBOSH = func(on bool) {
		fastFailBOSHMutex.Lock()
		defer fastFailBOSHMutex.Unlock()
		fastFailBOSH = on
	}

	var getFastFailBOSH = func() bool {
		fastFailBOSHMutex.Lock()
		defer fastFailBOSHMutex.Unlock()
		return fastFailBOSH
	}

	BeforeEach(func() {
		stdout = bytes.NewBuffer([]byte{})
		stderr = bytes.NewBuffer([]byte{})

		cmd = bosh.NewCmd(stdout, stderr, false)
		cmdWithDebug = bosh.NewCmd(stdout, stderr, true)

		fakeBOSHBackendServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			if getFastFailBOSH() {
				responseWriter.WriteHeader(http.StatusInternalServerError)
			}

			if request.Method == "POST" {
				boshArgsMutex.Lock()
				defer boshArgsMutex.Unlock()
				body, err := ioutil.ReadAll(request.Body)
				if err != nil {
					panic(err)
				}

				err = json.Unmarshal(body, &boshArgs)
				if err != nil {
					panic(err)
				}
			}
		}))

		var err error
		pathToFakeBOSH, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/fakebosh",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeBOSHBackendServer.URL))
		Expect(err).NotTo(HaveOccurred())

		pathToBOSH = filepath.Join(filepath.Dir(pathToFakeBOSH), "bosh")
		err = os.Rename(pathToFakeBOSH, pathToBOSH)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToBOSH), os.Getenv("PATH")}, ":"))
	})

	It("runs bosh with args", func() {
		err := cmd.Run("/tmp", []string{"create-env", "some-arg"})
		Expect(err).NotTo(HaveOccurred())

		boshArgsMutex.Lock()
		defer boshArgsMutex.Unlock()
		Expect(boshArgs).To(Equal([]string{"create-env", "some-arg"}))

		Expect(stdout).NotTo(MatchRegexp("working directory: (.*)/tmp"))
		Expect(stdout).NotTo(ContainSubstring("create-env some-arg"))
	})

	It("redirects command stdout to provided stdout when debug is true", func() {
		err := cmdWithDebug.Run("/tmp", []string{"create-env", "some-arg"})
		Expect(err).NotTo(HaveOccurred())

		Expect(stdout.String()).To(MatchRegexp("working directory: (.*)/tmp"))
		Expect(stdout.String()).To(ContainSubstring("create-env some-arg"))
	})

	Context("failure case", func() {
		BeforeEach(func() {
			setFastFailBOSH(true)
		})

		AfterEach(func() {
			setFastFailBOSH(false)
		})

		It("returns an error when terraform fails", func() {
			err := cmd.Run("", []string{})
			Expect(err).To(MatchError("exit status 1"))
		})

		It("redirects command stderr to provided stderr when debug is true", func() {
			_ = cmdWithDebug.Run("", []string{})
			Expect(stderr.String()).To(ContainSubstring("failed to bosh"))
		})
	})
})
