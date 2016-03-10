package storage_test

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Store", func() {
	var (
		store   storage.Store
		tempDir string
	)

	BeforeEach(func() {
		store = storage.NewStore()

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		storage.ResetEncode()
	})

	Describe("Set", func() {
		It("stores the state into a file", func() {
			err := store.Set(tempDir, storage.State{
				AWS: storage.AWS{
					AccessKeyID:     "some-aws-access-key-id",
					SecretAccessKey: "some-aws-secret-access-key",
					Region:          "some-region",
				},
				KeyPair: &storage.KeyPair{
					Name:       "some-name",
					PrivateKey: "some-private",
					PublicKey:  "some-public",
				},
				BOSH: &storage.BOSH{
					DirectorSSLCertificate: "some-bosh-ssl-certificate",
					DirectorSSLPrivateKey:  "some-bosh-ssl-private-key",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			data, err := ioutil.ReadFile(filepath.Join(tempDir, "state.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(MatchJSON(`{
				"version": 1,
				"aws": {
					"accessKeyId": "some-aws-access-key-id",
					"secretAccessKey": "some-aws-secret-access-key",
					"region": "some-region"
				},
				"keyPair": {
					"name": "some-name",
					"privateKey": "some-private",
					"publicKey": "some-public"
				},
				"bosh":{
					"directorSSLCertificate": "some-bosh-ssl-certificate",
					"directorSSLPrivateKey": "some-bosh-ssl-private-key"
				}
			}`))
		})

		Context("failure cases", func() {
			It("fails to open the state.json file", func() {
				err := os.Chmod(tempDir, 0000)
				Expect(err).NotTo(HaveOccurred())

				err = store.Set(tempDir, storage.State{})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})

			It("fails to write the state.json file", func() {
				storage.SetEncode(func(io.Writer, interface{}) error {
					return errors.New("failed to encode")
				})

				err := store.Set(tempDir, storage.State{})
				Expect(err).To(MatchError("failed to encode"))
			})
		})
	})

	Describe("Get", func() {
		It("gets the aws credentials", func() {
			err := ioutil.WriteFile(filepath.Join(tempDir, "state.json"), []byte(`{
				"version": 1,
				"aws": {
					"accessKeyId": "some-aws-access-key-id",
					"secretAccessKey": "some-aws-secret-access-key",
					"region": "some-aws-region"
				},
				"keyPair": {
					"name": "some-name",
					"privateKey": "some-private-key",
					"publicKey": "some-public-key"
				},
				"bosh": {
					"directorSSLCertificate": "some-bosh-ssl-certificate",
					"directorSSLPrivateKey": "some-bosh-ssl-private-key"
				}
			}`), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			state, err := store.Get(tempDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(state).To(Equal(storage.State{
				Version: 1,
				AWS: storage.AWS{
					AccessKeyID:     "some-aws-access-key-id",
					SecretAccessKey: "some-aws-secret-access-key",
					Region:          "some-aws-region",
				},
				KeyPair: &storage.KeyPair{
					Name:       "some-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
				BOSH: &storage.BOSH{
					DirectorSSLCertificate: "some-bosh-ssl-certificate",
					DirectorSSLPrivateKey:  "some-bosh-ssl-private-key",
				},
			}))
		})

		Context("when the state.json file doesn't exist", func() {
			It("returns an empty state object", func() {
				state, err := store.Get(tempDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(state).To(Equal(storage.State{}))
			})
		})

		Context("failure cases", func() {
			It("fails to open the state.json file", func() {
				err := os.Chmod(tempDir, 0000)
				Expect(err).NotTo(HaveOccurred())

				_, err = store.Get(tempDir)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})

			It("fails to decode the state.json file", func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "state.json"), []byte(`%%%%`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				_, err = store.Get(tempDir)
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})
		})
	})
})
