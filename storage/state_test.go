package storage_test

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/pivotal-cf-experimental/gomegamatchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Store", func() {
	var (
		store   storage.Store
		tempDir string
	)

	BeforeEach(func() {

		var err error
		tempDir, err = ioutil.TempDir("", "")

		store = storage.NewStore(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		storage.ResetEncode()
	})

	Describe("Set", func() {
		It("stores the state into a file", func() {
			err := store.Set(storage.State{
				IAAS: "aws",
				AWS: storage.AWS{
					AccessKeyID:     "some-aws-access-key-id",
					SecretAccessKey: "some-aws-secret-access-key",
					Region:          "some-region",
				},
				GCP: storage.GCP{
					ServiceAccountKey: "some-service-account-key",
					ProjectID:         "some-project-id",
					Zone:              "some-zone",
					Region:            "some-region",
				},
				KeyPair: storage.KeyPair{
					Name:       "some-name",
					PrivateKey: "some-private",
					PublicKey:  "some-public",
				},
				LB: storage.LB{
					Type: "some-type",
					Cert: "some-cert",
					Key:  "some-key",
				},
				BOSH: storage.BOSH{
					DirectorName:           "some-director-name",
					DirectorUsername:       "some-director-username",
					DirectorPassword:       "some-director-password",
					DirectorAddress:        "some-director-address",
					DirectorSSLCA:          "some-bosh-ssl-ca",
					DirectorSSLCertificate: "some-bosh-ssl-certificate",
					DirectorSSLPrivateKey:  "some-bosh-ssl-private-key",
					State: map[string]interface{}{
						"key": "value",
					},
					Manifest: "name: bosh",
					Credentials: map[string]string{
						"mbusUsername":              "some-mbus-username",
						"natsUsername":              "some-nats-username",
						"postgresUsername":          "some-postgres-username",
						"registryUsername":          "some-registry-username",
						"blobstoreDirectorUsername": "some-blobstore-director-username",
						"blobstoreAgentUsername":    "some-blobstore-agent-username",
						"hmUsername":                "some-hm-username",
						"mbusPassword":              "some-mbus-password",
						"natsPassword":              "some-nats-password",
						"postgresPassword":          "some-postgres-password",
						"registryPassword":          "some-registry-password",
						"blobstoreDirectorPassword": "some-blobstore-director-password",
						"blobstoreAgentPassword":    "some-blobstore-agent-password",
						"hmPassword":                "some-hm-password",
					},
				},
				Stack: storage.Stack{
					Name:            "some-stack-name",
					LBType:          "some-lb-type",
					CertificateName: "some-certificate-name",
				},
				EnvID:   "some-env-id",
				TFState: "some-tf-state",
			})
			Expect(err).NotTo(HaveOccurred())

			data, err := ioutil.ReadFile(filepath.Join(tempDir, "bbl-state.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(MatchJSON(`{
				"version": 2,
				"iaas": "aws",
				"aws": {
					"accessKeyId": "some-aws-access-key-id",
					"secretAccessKey": "some-aws-secret-access-key",
					"region": "some-region"
				},
				"gcp": {
					"serviceAccountKey": "some-service-account-key",
					"projectID": "some-project-id",
					"zone": "some-zone",
					"region": "some-region"
				},
				"keyPair": {
					"name": "some-name",
					"privateKey": "some-private",
					"publicKey": "some-public"
				},
				"lb": {
					"type": "some-type",
					"cert": "some-cert",
					"key": "some-key"
				},
				"bosh":{
					"directorName": "some-director-name",
					"directorUsername": "some-director-username",
					"directorPassword": "some-director-password",
					"directorAddress": "some-director-address",
					"directorSSLCA": "some-bosh-ssl-ca",
					"directorSSLCertificate": "some-bosh-ssl-certificate",
					"directorSSLPrivateKey": "some-bosh-ssl-private-key",
					"credentials": {
						"mbusUsername": "some-mbus-username",
						"natsUsername": "some-nats-username",
						"postgresUsername": "some-postgres-username",
						"registryUsername": "some-registry-username",
						"blobstoreDirectorUsername": "some-blobstore-director-username",
						"blobstoreAgentUsername": "some-blobstore-agent-username",
						"hmUsername": "some-hm-username",
						"mbusPassword": "some-mbus-password",
						"natsPassword": "some-nats-password",
						"postgresPassword": "some-postgres-password",
						"registryPassword": "some-registry-password",
						"blobstoreDirectorPassword": "some-blobstore-director-password",
						"blobstoreAgentPassword": "some-blobstore-agent-password",
						"hmPassword": "some-hm-password"
					},
					"manifest": "name: bosh",
					"state": {
						"key": "value"
					}
				},
				"stack": {
					"name": "some-stack-name",
					"lbType": "some-lb-type",
					"certificateName": "some-certificate-name"
				},
				"envID": "some-env-id",
				"tfState": "some-tf-state"
			}`))

			fileInfo, err := os.Stat(filepath.Join(tempDir, "bbl-state.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(fileInfo.Mode()).To(Equal(os.FileMode(0644)))
		})

		Context("when the state is empty", func() {
			It("removes the bbl-state.json file", func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte("{}"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = store.Set(storage.State{})
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Stat(filepath.Join(tempDir, "bbl-state.json"))
				Expect(os.IsNotExist(err)).To(BeTrue())
			})

			Context("when the bbl-state.json file does not exist", func() {
				It("does nothing", func() {
					err := store.Set(storage.State{})
					Expect(err).NotTo(HaveOccurred())

					_, err = os.Stat(filepath.Join(tempDir, "bbl-state.json"))
					Expect(os.IsNotExist(err)).To(BeTrue())
				})
			})

			Context("failure cases", func() {
				Context("when the bbl-state.json file cannot be removed", func() {
					It("returns an error", func() {
						err := os.Chmod(tempDir, 0000)
						Expect(err).NotTo(HaveOccurred())

						err = store.Set(storage.State{})
						Expect(err).To(MatchError(ContainSubstring("permission denied")))
					})
				})
			})
		})

		Context("failure cases", func() {
			It("fails when the directory does not exist", func() {
				store = storage.NewStore("non-valid-dir")
				err := store.Set(storage.State{})
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})

			It("fails to open the bbl-state.json file", func() {
				err := os.Chmod(tempDir, 0000)
				Expect(err).NotTo(HaveOccurred())

				err = store.Set(storage.State{
					Stack: storage.Stack{
						Name: "some-stack-name",
					},
				})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})

			It("fails to write the bbl-state.json file", func() {
				storage.SetEncode(func(io.Writer, interface{}) error {
					return errors.New("failed to encode")
				})

				err := store.Set(storage.State{
					Stack: storage.Stack{
						Name: "some-stack-name",
					},
				})
				Expect(err).To(MatchError("failed to encode"))
			})
		})
	})

	Describe("GCP", func() {
		Describe("Empty", func() {
			It("returns true when all fields are blank", func() {
				gcp := storage.GCP{}
				empty := gcp.Empty()
				Expect(empty).To(BeTrue())
			})

			It("returns false when at least one field is present", func() {
				gcp := storage.GCP{ServiceAccountKey: "some-account-key"}
				empty := gcp.Empty()
				Expect(empty).To(BeFalse())
			})
		})
	})

	Describe("GetState", func() {
		var logger *fakes.Logger

		BeforeEach(func() {
			logger = &fakes.Logger{}
			storage.GetStateLogger = logger
		})

		Context("when there is a v2 state file", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`{
					"version": 2,
					"iaas": "aws",
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
						"directorAddress": "some-director-address",
						"directorSSLCA": "some-bosh-ssl-ca",
						"directorSSLCertificate": "some-bosh-ssl-certificate",
						"directorSSLPrivateKey": "some-bosh-ssl-private-key",
						"manifest": "name: bosh"
					},
					"stack": {
						"name": "some-stack-name",
						"lbType": "some-lb",
						"certificateName": "some-certificate-name"
					}
				}`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the stored state information", func() {
				state, err := storage.GetState(tempDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(state).To(Equal(storage.State{
					Version: 2,
					IAAS:    "aws",
					AWS: storage.AWS{
						AccessKeyID:     "some-aws-access-key-id",
						SecretAccessKey: "some-aws-secret-access-key",
						Region:          "some-aws-region",
					},
					KeyPair: storage.KeyPair{
						Name:       "some-name",
						PrivateKey: "some-private-key",
						PublicKey:  "some-public-key",
					},
					BOSH: storage.BOSH{
						DirectorAddress:        "some-director-address",
						DirectorSSLCA:          "some-bosh-ssl-ca",
						DirectorSSLCertificate: "some-bosh-ssl-certificate",
						DirectorSSLPrivateKey:  "some-bosh-ssl-private-key",
						Manifest:               "name: bosh",
					},
					Stack: storage.Stack{
						Name:            "some-stack-name",
						LBType:          "some-lb",
						CertificateName: "some-certificate-name",
					},
				}))
			})
		})

		Context("when there is a v1 state file", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`{
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
						"directorAddress": "some-director-address",
						"directorSSLCA": "some-bosh-ssl-ca",
						"directorSSLCertificate": "some-bosh-ssl-certificate",
						"directorSSLPrivateKey": "some-bosh-ssl-private-key",
						"manifest": "name: bosh"
					},
					"stack": {
						"name": "some-stack-name",
						"lbType": "some-lb",
						"certificateName": "some-certificate-name"
					}
				}`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("migrates the state file with a default of iaas: aws", func() {
				state, err := storage.GetState(tempDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(state).To(Equal(storage.State{
					Version: 2,
					IAAS:    "aws",
					AWS: storage.AWS{
						AccessKeyID:     "some-aws-access-key-id",
						SecretAccessKey: "some-aws-secret-access-key",
						Region:          "some-aws-region",
					},
					KeyPair: storage.KeyPair{
						Name:       "some-name",
						PrivateKey: "some-private-key",
						PublicKey:  "some-public-key",
					},
					BOSH: storage.BOSH{
						DirectorAddress:        "some-director-address",
						DirectorSSLCA:          "some-bosh-ssl-ca",
						DirectorSSLCertificate: "some-bosh-ssl-certificate",
						DirectorSSLPrivateKey:  "some-bosh-ssl-private-key",
						Manifest:               "name: bosh",
					},
					Stack: storage.Stack{
						Name:            "some-stack-name",
						LBType:          "some-lb",
						CertificateName: "some-certificate-name",
					},
				}))
			})
		})

		Context("when the bbl-state.json file doesn't exist", func() {
			It("returns an empty state object", func() {
				state, err := storage.GetState(tempDir)
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
					}`), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					storage.ResetRename()
				})

				It("renames state.json to bbl-state.json and returns its state", func() {
					state, err := storage.GetState(tempDir)
					Expect(err).NotTo(HaveOccurred())

					_, err = os.Stat(filepath.Join(tempDir, "state.json"))
					Expect(err).To(gomegamatchers.BeAnOsIsNotExistError())

					_, err = os.Stat(filepath.Join(tempDir, "bbl-state.json"))
					Expect(err).NotTo(HaveOccurred())

					Expect(state).To(Equal(storage.State{
						Version: 2,
						AWS: storage.AWS{
							AccessKeyID:     "some-aws-access-key-id",
							SecretAccessKey: "some-aws-secret-access-key",
							Region:          "some-aws-region",
						},
					}))
					Expect(logger.PrintlnCall.CallCount).To(Equal(1))
					Expect(logger.PrintlnCall.Receives.Message).To(Equal("renaming state.json to bbl-state.json"))
				})

				Context("failure cases", func() {
					Context("when checking if state file exists fails", func() {
						It("returns an error", func() {
							err := os.Chmod(tempDir, os.FileMode(0000))
							Expect(err).NotTo(HaveOccurred())

							_, err = storage.GetState(tempDir)
							Expect(err).To(MatchError(ContainSubstring("permission denied")))
						})
					})

					Context("when renaming the file fails", func() {
						It("returns an error", func() {
							storage.SetRename(func(src, dst string) error {
								return errors.New("renaming failed")
							})

							_, err := storage.GetState(tempDir)
							Expect(err).To(MatchError("renaming failed"))
						})
					})
				})
			})
		})

		Context("failure cases", func() {
			It("fails when the directory does not exist", func() {
				_, err := storage.GetState("some-fake-directory")
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})

			It("fails to open the bbl-state.json file", func() {
				err := os.Chmod(tempDir, 0000)
				Expect(err).NotTo(HaveOccurred())

				_, err = storage.GetState(tempDir)
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})

			It("fails to decode the bbl-state.json file", func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`%%%%`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				_, err = storage.GetState(tempDir)
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
			})

			It("returns error when bbl-state.json and state.json exists", func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "state.json"), []byte(`{}`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`{}`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				_, err = storage.GetState(tempDir)
				Expect(err).To(MatchError(ContainSubstring("Cannot proceed with state.json and bbl-state.json present. Please delete one of the files.")))

			})
		})
	})
})
