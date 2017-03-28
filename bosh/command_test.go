package bosh_test

import (
	"bytes"
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

		cmd bosh.Cmd

		fakeBOSHBackendServer *httptest.Server
		pathToFakeBOSH        string
		pathToBOSH            string

		fastFailBOSH      bool
		fastFailBOSHMutex sync.Mutex

		boshArgs      string
		boshArgsMutex sync.Mutex

		tempDir string
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

		cmd = bosh.NewCmd(stderr)

		fakeBOSHBackendServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			switch request.URL.Path {
			case "/create-env/args":
				boshArgsMutex.Lock()
				defer boshArgsMutex.Unlock()
				body, err := ioutil.ReadAll(request.Body)
				Expect(err).NotTo(HaveOccurred())
				boshArgs = string(body)
			case "/create-env/fastfail":
				if getFastFailBOSH() {
					responseWriter.WriteHeader(http.StatusInternalServerError)
				} else {
					responseWriter.WriteHeader(http.StatusOK)
				}
				return
			default:
				responseWriter.WriteHeader(http.StatusOK)
			}
		}))

		var err error
		pathToFakeBOSH, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl/fakebosh",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeBOSHBackendServer.URL))
		Expect(err).NotTo(HaveOccurred())

		pathToBOSH = filepath.Join(filepath.Dir(pathToFakeBOSH), "bosh")
		err = os.Rename(pathToFakeBOSH, pathToBOSH)
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("PATH", strings.Join([]string{filepath.Dir(pathToBOSH), originalPath}, ":"))

		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.Setenv("PATH", originalPath)
	})

	It("runs bosh with args", func() {
		err := cmd.Run(stdout, tempDir, []string{"create-env", "some-arg"})
		Expect(err).NotTo(HaveOccurred())

		boshArgsMutex.Lock()
		defer boshArgsMutex.Unlock()
		Expect(boshArgs).To(Equal(`["create-env","some-arg"]`))

		Expect(stdout).To(MatchRegexp(fmt.Sprintf("working directory: (.*)%s", tempDir)))
		Expect(stdout).To(ContainSubstring("create-env some-arg"))
	})

	Context("failure case", func() {
		BeforeEach(func() {
			setFastFailBOSH(true)
		})

		AfterEach(func() {
			setFastFailBOSH(false)
		})

		It("returns an error when bosh fails", func() {
			err := cmd.Run(stdout, tempDir, []string{"create-env"})
			Expect(err).To(MatchError("exit status 1"))
			Expect(stderr.String()).To(ContainSubstring("failed to bosh"))
		})
	})
})
