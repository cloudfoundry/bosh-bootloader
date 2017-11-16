package storage_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StateBootstrap", func() {
	Describe("GetState", func() {
		var (
			logger        *fakes.Logger
			bootstrap     storage.StateBootstrap
			tempDir       string
			latestVersion string
		)

		BeforeEach(func() {
			logger = &fakes.Logger{}
			latestVersion = "latest"
			bootstrap = storage.NewStateBootstrap(logger, latestVersion)

			var err error
			tempDir, err = ioutil.TempDir("", "")

			Expect(err).NotTo(HaveOccurred())
		})

		Context("when there is a completely empty state file", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`{}`), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a new state", func() {
				state, err := bootstrap.GetState(tempDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(state).To(Equal(storage.State{
					Version:    13,
					BBLVersion: latestVersion,
				}))
			})
		})

		Context("when there is a pre v3 state file", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`{
					"version": 2
				}`), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := bootstrap.GetState(tempDir)
				Expect(err).To(MatchError("Existing bbl environment is incompatible with bbl v3. Create a new environment with v3 to continue."))
			})
		})

		Context("when there is a current version state file", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`{
					"version": 12,
					"bblVersion": "some-bbl-version",
					"iaas": "aws",
					"aws": {
						"region": "some-aws-region"
					},
					"bosh": {
						"directorAddress": "some-director-address",
						"directorSSLCA": "some-bosh-ssl-ca",
						"directorSSLCertificate": "some-bosh-ssl-certificate",
						"directorSSLPrivateKey": "some-bosh-ssl-private-key",
						"manifest": "name: bosh"
					}
				}`), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the stored state information", func() {
				state, err := bootstrap.GetState(tempDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(state).To(Equal(storage.State{
					Version:    12,
					BBLVersion: "some-bbl-version",
					IAAS:       "aws",
					AWS: storage.AWS{
						Region: "some-aws-region",
					},
					BOSH: storage.BOSH{
						DirectorAddress:        "some-director-address",
						DirectorSSLCA:          "some-bosh-ssl-ca",
						DirectorSSLCertificate: "some-bosh-ssl-certificate",
						DirectorSSLPrivateKey:  "some-bosh-ssl-private-key",
						Manifest:               "name: bosh",
					},
				}))
			})
		})

		Context("when there is a state file missing BBL version", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`{
					"version": 12
				}`), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())
			})

			It("populates BBL version based on state version", func() {
				state, err := bootstrap.GetState(tempDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(state).To(Equal(storage.State{
					Version:    12,
					BBLVersion: "5.1.0",
				}))
			})
		})

		Context("when there is a state file with a newer version than internal version", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`{
					"version": 999,
					"bblVersion": "9.9.9"
				}`), storage.StateMode)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := bootstrap.GetState(tempDir)
				Expect(err).To(MatchError("Existing bbl environment was created with a newer version of bbl. Please upgrade to bbl v9.9.9.\n"))
			})
		})

		Context("when the bbl-state.json file doesn't exist", func() {
			It("returns an empty state object", func() {
				state, err := bootstrap.GetState(tempDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(state).To(Equal(storage.State{}))
			})

			Context("when state.json exists", func() {
				BeforeEach(func() {
					err := ioutil.WriteFile(filepath.Join(tempDir, "state.json"), []byte(`{
						"version": 2,
						"aws": {
							"accessKeyId": "some-aws-access-key-id",
							"secretAccessKey": "some-aws-secret-access-key",
							"region": "some-aws-region"
						}
					}`), storage.StateMode)
					Expect(err).NotTo(HaveOccurred())
				})

				Context("failure cases", func() {
					Context("when checking if state file exists fails", func() {
						It("returns an error", func() {
							err := os.Chmod(tempDir, os.FileMode(0000))
							Expect(err).NotTo(HaveOccurred())

							_, err = bootstrap.GetState(tempDir)
							Expect(err).To(MatchError(ContainSubstring("permission denied")))
						})
					})
				})
			})
		})

		Context("failure cases", func() {
			Context("when the directory does not exist", func() {
				It("returns an error", func() {
					_, err := bootstrap.GetState("some-fake-directory")
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			Context("when it fails to open the bbl-state.json file", func() {
				It("returns an error", func() {
					err := os.Chmod(tempDir, 0000)
					Expect(err).NotTo(HaveOccurred())

					_, err = bootstrap.GetState(tempDir)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			Context("when it fails to decode the bbl-state.json file", func() {
				It("returns an error", func() {
					err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`%%%%`), storage.StateMode)
					Expect(err).NotTo(HaveOccurred())

					_, err = bootstrap.GetState(tempDir)
					Expect(err).To(MatchError(ContainSubstring("invalid character")))
				})
			})
		})
	})
})
