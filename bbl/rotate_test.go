package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bbl/awsbackend"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rosenhouse/awsfaker"
)

var _ = Describe("bbl rotate", func() {
	Context("when iaas is aws", func() {
		var (
			tempDirectory string
			stateContents []byte

			fakeAWS        *awsbackend.Backend
			fakeAWSServer  *httptest.Server
			fakeBOSHServer *httptest.Server
			fakeBOSH       *fakeBOSHDirector
		)

		BeforeEach(func() {
			var err error
			tempDirectory, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			stateContents = []byte(`{
				"version": 3,
				"iaas": "aws",
				"aws": {
					"accessKeyId": "some-access-key-id",
					"secretAccessKey": "some-secret-access-key",
					"region": "some-region"
				},
				"keyPair": {
					"privateKey": "some-private-key",
					"publicKey": "some-public-key"
				},
				"stack": {
					"name": "some-stack-name"
				}
			}`)

			fakeBOSH = &fakeBOSHDirector{}
			fakeBOSHServer = httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				fakeBOSH.ServeHTTP(responseWriter, request)
			}))

			fakeAWS = awsbackend.New(fakeBOSHServer.URL)
			fakeAWSServer = httptest.NewServer(awsfaker.New(fakeAWS))

			fakeAWS.Stacks.Set(awsbackend.Stack{
				Name: "some-stack-name",
			})
		})

		It("rotates the keys of the director", func() {
			err := ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), stateContents, os.ModePerm)

			Expect(err).NotTo(HaveOccurred())
			args := []string{
				fmt.Sprintf("--endpoint-override=%s", fakeAWSServer.URL),
				"--state-dir", tempDirectory,
				"rotate",
			}

			executeCommand(args, 0)

			newState := readStateJson(tempDirectory)
			Expect(newState.KeyPair.PrivateKey).To(MatchRegexp(`-----BEGIN RSA PRIVATE KEY-----((.|\n)*)-----END RSA PRIVATE KEY-----`))
		})
	})

	Context("when iaas is gcp", func() {
		var (
			tempDirectory string
			stateContents []byte
		)

		BeforeEach(func() {
			var err error
			tempDirectory, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			stateContents = []byte(fmt.Sprintf(`{
				"version": 3,
				"iaas": "gcp",
				"gcp": {
					"serviceAccountKey": %q,
					"projectID": "some-project-id",
					"zone": "some-zone",
					"region": "some-region"
				},
				"keyPair": {
					"privateKey": "some-private-key",
					"publicKey": "some-public-key"
				},
				"noDirector": true
			}`, serviceAccountKey))
		})

		It("rotates the keys", func() {
			err := ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), stateContents, os.ModePerm)

			Expect(err).NotTo(HaveOccurred())
			args := []string{
				"--state-dir", tempDirectory,
				"rotate",
			}

			executeCommand(args, 0)

			newState := readStateJson(tempDirectory)
			Expect(newState.KeyPair.PrivateKey).To(MatchRegexp(`-----BEGIN RSA PRIVATE KEY-----((.|\n)*)-----END RSA PRIVATE KEY-----`))
			Expect(newState.KeyPair.PublicKey).To(HavePrefix("ssh-rsa"))
		})
	})
})
