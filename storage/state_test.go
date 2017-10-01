package storage_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	uuid "github.com/nu7hatch/gouuid"

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
		storage.ResetUUIDNewV4()
		storage.ResetMarshalIndent()
	})

	Describe("Set", func() {
		Context("when credhub is enabled", func() {
			It("stores the state into a file, without IAAS credentials", func() {
				storage.SetUUIDNewV4(func() (*uuid.UUID, error) {
					return &uuid.UUID{
						0x01, 0x02, 0x03, 0x04,
						0x05, 0x06, 0x07, 0x08,
						0x09, 0x10, 0x11, 0x12,
						0x13, 0x14, 0x15, 0x16}, nil
				})
				err := store.Set(storage.State{
					IAAS: "aws",
					AWS: storage.AWS{
						AccessKeyID:     "some-aws-access-key-id",
						SecretAccessKey: "some-aws-secret-access-key",
						Region:          "some-region",
					},
					Azure: storage.Azure{
						ClientID:       "client-id",
						ClientSecret:   "client-secret",
						Location:       "location",
						SubscriptionID: "subscription-id",
						TenantID:       "tenant-id",
					},
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						ProjectID:         "some-project-id",
						Zone:              "some-zone",
						Region:            "some-region",
						Zones:             []string{"some-zone", "some-other-zone"},
					},
					KeyPair: storage.KeyPair{
						Name:       "some-name",
						PrivateKey: "some-private",
						PublicKey:  "some-public",
					},
					LB: storage.LB{
						Type:   "some-type",
						Cert:   "some-cert",
						Key:    "some-key",
						Chain:  "some-chain",
						Domain: "some-domain",
					},
					Jumpbox: storage.Jumpbox{
						URL:       "some-jumpbox-url",
						Manifest:  "name: jumpbox",
						Variables: "some-jumpbox-vars",
						State: map[string]interface{}{
							"key": "value",
						},
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
						Variables:   "some-vars",
						Manifest:    "name: bosh",
						UserOpsFile: "some-ops-file",
					},
					EnvID:   "some-env-id",
					TFState: "some-tf-state",
				})
				Expect(err).NotTo(HaveOccurred())

				data, err := ioutil.ReadFile(filepath.Join(tempDir, "bbl-state.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(MatchJSON(`{
				"version": 11,
				"iaas": "aws",
				"noDirector": false,
				"aws": {
					"region": "some-region"
				},
				"azure": {
					"clientId": "client-id",
					"clientSecret": "client-secret",
					"location": "location",
					"subscriptionId": "subscription-id",
					"tenantId": "tenant-id"
				},
				"gcp": {
					"zone": "some-zone",
					"region": "some-region",
					"zones": ["some-zone", "some-other-zone"]
				},
				"keyPair": {
					"name": "some-name",
					"privateKey": "some-private",
					"publicKey": "some-public"
				},
				"lb": {
					"type": "some-type",
					"cert": "some-cert",
					"key": "some-key",
					"chain": "some-chain",
					"domain": "some-domain"
				},
				"jumpbox":{
					"url": "some-jumpbox-url",
					"variables": "some-jumpbox-vars",
					"manifest": "name: jumpbox",
					"state": {
						"key": "value"
					}
				},
				"bosh":{
					"directorName": "some-director-name",
					"directorUsername": "some-director-username",
					"directorPassword": "some-director-password",
					"directorAddress": "some-director-address",
					"directorSSLCA": "some-bosh-ssl-ca",
					"directorSSLCertificate": "some-bosh-ssl-certificate",
					"directorSSLPrivateKey": "some-bosh-ssl-private-key",
					"variables":   "some-vars",
					"manifest": "name: bosh",
					"userOpsFile": "some-ops-file",
					"state": {
						"key": "value"
					}
				},
				"envID": "some-env-id",
				"tfState": "some-tf-state",
				"id": "01020304-0506-0708-0910-111213141516",
				"latestTFOutput": ""
		    	}`))

				fileInfo, err := os.Stat(filepath.Join(tempDir, "bbl-state.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(fileInfo.Mode()).To(Equal(os.FileMode(0644)))
			})
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

				Context("when uuid new V4 fails", func() {
					It("returns an error", func() {
						storage.SetUUIDNewV4(func() (*uuid.UUID, error) {
							return nil, errors.New("some error")
						})
						err := store.Set(storage.State{
							IAAS: "some-iaas",
						})
						Expect(err).To(MatchError("Create state ID: some error"))
					})
				})
			})
		})

		Context("failure cases", func() {
			Context("when json marshalling fails", func() {
				BeforeEach(func() {
					storage.SetMarshalIndent(func(state interface{}, prefix string, indent string) ([]byte, error) {
						return []byte{}, errors.New("failed to marshal JSON")
					})
				})

				It("returns an error", func() {
					err := store.Set(storage.State{
						IAAS: "aws",
					})
					Expect(err).To(MatchError("failed to marshal JSON"))
				})
			})

			Context("when the directory does not exist", func() {
				BeforeEach(func() {
					storage.SetMarshalIndent(func(state interface{}, prefix string, indent string) ([]byte, error) {
						return []byte{}, errors.New("failed to marshal JSON")
					})
				})

				It("returns an error", func() {
					store = storage.NewStore("non-valid-dir")
					err := store.Set(storage.State{})
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			Context("when it fails to open the bbl-state.json file", func() {
				BeforeEach(func() {
					err := os.Chmod(tempDir, 0000)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					err := store.Set(storage.State{})
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})

	Describe("GCP", func() {
		Describe("Empty", func() {
			var gcp storage.GCP
			Context("when all fields are blank", func() {
				BeforeEach(func() {
					gcp = storage.GCP{}
				})

				It("returns true", func() {
					empty := gcp.Empty()
					Expect(empty).To(BeTrue())
				})
			})

			Context("when at least one field is present", func() {
				BeforeEach(func() {
					gcp = storage.GCP{ServiceAccountKey: "some-account-key"}
				})

				It("returns false", func() {
					empty := gcp.Empty()
					Expect(empty).To(BeFalse())
				})
			})
		})
	})

	Describe("GetState", func() {
		var logger *fakes.Logger

		BeforeEach(func() {
			logger = &fakes.Logger{}
			storage.GetStateLogger = logger
		})

		Context("when there is a completely empty state file", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`{}`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a new state", func() {
				state, err := storage.GetState(tempDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(state).To(Equal(storage.State{
					Version: 11,
				}))
			})
		})

		Context("when there is a pre v3 state file", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`{
					"version": 2
				}`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := storage.GetState(tempDir)
				Expect(err).To(MatchError("Existing bbl environment is incompatible with bbl v3. Create a new environment with v3 to continue."))
			})
		})

		Context("when there is a v11 state file", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`{
					"version": 11,
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
					}
				}`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns the stored state information", func() {
				state, err := storage.GetState(tempDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(state).To(Equal(storage.State{
					Version: 11,
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
				}))
			})
		})

		Context("when there is a state file with a newer version than internal version", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`{
					"version": 9999
				}`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				_, err := storage.GetState(tempDir)
				Expect(err).To(MatchError("Existing bbl environment was created with a newer version of bbl. Please upgrade to a version of bbl compatible with schema version 9999.\n"))
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

				Context("failure cases", func() {
					Context("when checking if state file exists fails", func() {
						It("returns an error", func() {
							err := os.Chmod(tempDir, os.FileMode(0000))
							Expect(err).NotTo(HaveOccurred())

							_, err = storage.GetState(tempDir)
							Expect(err).To(MatchError(ContainSubstring("permission denied")))
						})
					})
				})
			})
		})

		Context("failure cases", func() {
			Context("when the directory does not exist", func() {
				It("returns an error", func() {
					_, err := storage.GetState("some-fake-directory")
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			Context("when it fails to open the bbl-state.json file", func() {
				It("returns an error", func() {
					err := os.Chmod(tempDir, 0000)
					Expect(err).NotTo(HaveOccurred())

					_, err = storage.GetState(tempDir)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			Context("when it fails to decode the bbl-state.json file", func() {
				It("returns an error", func() {
					err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`%%%%`), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					_, err = storage.GetState(tempDir)
					Expect(err).To(MatchError(ContainSubstring("invalid character")))
				})
			})
		})
	})
})
