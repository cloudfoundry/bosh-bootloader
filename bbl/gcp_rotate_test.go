package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bbl rotate gcp", func() {
	var (
		tempDirectory string
	)

	BeforeEach(func() {
		var err error
		tempDirectory, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	It("rotates the keys of the director", func() {
		state := []byte(fmt.Sprintf(`{
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
			}
		}`, serviceAccountKey))
		err := ioutil.WriteFile(filepath.Join(tempDirectory, storage.StateFileName), state, os.ModePerm)

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
