package storage_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

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
		storage.ResetMarshalIndent()
	})

	Describe("Set", func() {
		Context("when credhub is enabled", func() {
			It("stores the state into a file, sans IAAS credentials", func() {
				err := store.Set(storage.State{
					IAAS: "aws",
					AWS: storage.AWS{
						AccessKeyID:     "some-aws-access-key-id",
						SecretAccessKey: "some-aws-secret-access-key",
						Region:          "some-region",
					},
					Azure: storage.Azure{
						SubscriptionID: "subscription-id",
						TenantID:       "tenant-id",
						ClientID:       "client-id",
						ClientSecret:   "client-secret",
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
						Enabled:   true,
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
						BOSHAZ:          "some-bosh-az",
					},
					EnvID:   "some-env-id",
					TFState: "some-tf-state",
				})
				Expect(err).NotTo(HaveOccurred())

				data, err := ioutil.ReadFile(filepath.Join(tempDir, "bbl-state.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(MatchJSON(`{
				"version": 8,
				"iaas": "aws",
				"noDirector": false,
				"migratedFromCloudFormation": false,
				"jumpbox": true,
				"aws": {
					"region": "some-region"
				},
				"azure": {
					"subscriptionId": "subscription-id",
					"tenantId": "tenant-id",
					"clientId": "client-id",
					"clientSecret": "client-secret"
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
					"enabled": true,
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
					"variables":   "some-vars",
					"manifest": "name: bosh",
					"userOpsFile": "some-ops-file",
					"state": {
						"key": "value"
					}
				},
				"stack": {
					"name": "some-stack-name",
					"lbType": "some-lb-type",
					"certificateName": "some-certificate-name",
					"boshAZ": "some-bosh-az"
				},
				"envID": "some-env-id",
				"tfState": "some-tf-state",
				"latestTFOutput": ""
		    	}`))

				fileInfo, err := os.Stat(filepath.Join(tempDir, "bbl-state.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(fileInfo.Mode()).To(Equal(os.FileMode(0644)))
			})
		})

		Context("when --credhub is not enabled", func() {
			It("persists IAAS credentials", func() {
				err := store.Set(storage.State{
					IAAS: "aws",
					AWS: storage.AWS{
						AccessKeyID:     "some-aws-access-key-id",
						SecretAccessKey: "some-aws-secret-access-key",
						Region:          "some-region",
					},
					Azure: storage.Azure{
						SubscriptionID: "subscription-id",
						TenantID:       "tenant-id",
						ClientID:       "client-id",
						ClientSecret:   "client-secret",
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
						Enabled: false,
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
						BOSHAZ:          "some-bosh-az",
					},
					EnvID:   "some-env-id",
					TFState: "some-tf-state",
				})
				Expect(err).NotTo(HaveOccurred())

				data, err := ioutil.ReadFile(filepath.Join(tempDir, "bbl-state.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(MatchJSON(`{
					"version": 8,
					"iaas": "aws",
					"noDirector": false,
					"migratedFromCloudFormation": false,
					"aws": {
						"accessKeyId": "some-aws-access-key-id",
						"secretAccessKey": "some-aws-secret-access-key",
						"region": "some-region"
					},
					"azure": {
						"subscriptionId": "subscription-id",
						"tenantId": "tenant-id",
						"clientId": "client-id",
						"clientSecret": "client-secret"
					},
					"gcp": {
						"serviceAccountKey": "some-service-account-key",
						"projectID": "some-project-id",
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
						"enabled": false,
						"url": "",
						"variables": "",
						"manifest": "",
						"state": null
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
						"variables":   "some-vars",
						"manifest": "name: bosh",
						"userOpsFile": "some-ops-file",
						"state": {
							"key": "value"
						}
					},
					"stack": {
						"name": "some-stack-name",
						"lbType": "some-lb-type",
						"certificateName": "some-certificate-name",
						"boshAZ": "some-bosh-az"
					},
					"envID": "some-env-id",
					"tfState": "some-tf-state",
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
			})
		})

		Context("failure cases", func() {
			It("fails when json marshalling doesn't work", func() {
				storage.SetMarshalIndent(func(state interface{}, prefix string, indent string) ([]byte, error) {
					return []byte{}, errors.New("failed to marshal JSON")
				})
				err := store.Set(storage.State{
					IAAS: "aws",
				})
				Expect(err).To(MatchError("failed to marshal JSON"))
			})

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

		Context("when there is a completely empty state file", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`{}`), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns a new state", func() {
				state, err := storage.GetState(tempDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(state).To(Equal(storage.State{
					Version: 8,
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

		Context("when there is a v8 state file", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(filepath.Join(tempDir, "bbl-state.json"), []byte(`{
					"version": 8,
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
					Version: 8,
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
		})
	})
})
