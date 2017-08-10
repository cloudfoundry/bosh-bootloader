package bosh_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
		pathToBOSH            string

		fastFailBOSH      bool
		fastFailBOSHMutex sync.Mutex

		boshArgs      string
		boshArgsMutex sync.Mutex

		tempDir string
	)

	var setFastFailBOSH = func(on bool) {
		fastFailBOSHMutex.Lock()

		fastFailBOSH = on

		fastFailBOSHMutex.Unlock()
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

				body, err := ioutil.ReadAll(request.Body)
				Expect(err).NotTo(HaveOccurred())
				boshArgs = string(body)

				boshArgsMutex.Unlock()
			case "/create-env/fastfail":
				if getFastFailBOSH() {
					responseWriter.WriteHeader(http.StatusInternalServerError)
					return
				}

				responseWriter.WriteHeader(http.StatusOK)
			default:
				responseWriter.WriteHeader(http.StatusOK)
			}
		}))

		var err error
		pathToBOSH, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/fakes/bosh",
			"--ldflags", fmt.Sprintf("-X main.backendURL=%s", fakeBOSHBackendServer.URL))
		Expect(err).NotTo(HaveOccurred())

		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.Setenv("PATH", originalPath)
		gexec.CleanupBuildArtifacts()
	})

	Context("when a user has bosh", func() {
		It("runs bosh with args", func() {
			os.Setenv("PATH", filepath.Dir(pathToBOSH))

			err := cmd.Run(stdout, tempDir, []string{"create-env", "some-arg"})
			Expect(err).NotTo(HaveOccurred())

			boshArgsMutex.Lock()
			defer boshArgsMutex.Unlock()
			Expect(boshArgs).To(Equal(`["create-env","some-arg"]`))

			Expect(stdout).To(MatchRegexp(fmt.Sprintf("working directory: (.*)%s", tempDir)))
			Expect(stdout).To(ContainSubstring("create-env some-arg"))
		})
	})

	Context("when a user has bosh2", func() {
		It("runs bosh2 with args", func() {
			err := os.Rename(pathToBOSH, filepath.Join(filepath.Dir(pathToBOSH), "bosh2"))
			Expect(err).NotTo(HaveOccurred())

			bosh2 := filepath.Join(filepath.Dir(pathToBOSH), "bosh2")
			err = os.Setenv("PATH", filepath.Dir(bosh2))
			Expect(err).NotTo(HaveOccurred())

			err = cmd.Run(stdout, tempDir, []string{"create-env", "some-arg"})
			Expect(err).NotTo(HaveOccurred())

			boshArgsMutex.Lock()
			defer boshArgsMutex.Unlock()
			Expect(boshArgs).To(Equal(`["create-env","some-arg"]`))

			Expect(stdout).To(MatchRegexp(fmt.Sprintf("working directory: (.*)%s", tempDir)))
			Expect(stdout).To(ContainSubstring("create-env some-arg"))
		})
	})

	Context("when an error occurs", func() {
		BeforeEach(func() {
			setFastFailBOSH(true)
		})

		AfterEach(func() {
			setFastFailBOSH(false)
		})

		Context("when bosh fails", func() {
			It("returns an error", func() {
				os.Setenv("PATH", filepath.Dir(pathToBOSH))

				err := cmd.Run(stdout, tempDir, []string{"create-env"})
				Expect(err).To(MatchError("exit status 1"))
				Expect(stderr.String()).To(ContainSubstring("failed to bosh"))
			})
		})
	})
})
