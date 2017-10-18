package storage_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	uuid "github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
				"version": 12,
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

			It("removes create-env scripts", func() {
				createDirector := filepath.Join(tempDir, "create-director.sh")
				createJumpbox := filepath.Join(tempDir, "create-jumpbox.sh")
				err := ioutil.WriteFile(createDirector, []byte("#!/bin/bash"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
				err = ioutil.WriteFile(createJumpbox, []byte("#!/bin/bash"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = store.Set(storage.State{})
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Stat(createDirector)
				Expect(os.IsNotExist(err)).To(BeTrue())
				_, err = os.Stat(createJumpbox)
				Expect(os.IsNotExist(err)).To(BeTrue())
			})

			DescribeTable("removing bbl-created directories",
				func(directory string, expectToBeDeleted bool) {
					err := os.Mkdir(filepath.Join(tempDir, directory), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					err = ioutil.WriteFile(filepath.Join(tempDir, directory, "foo.txt"), []byte("{}"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					err = store.Set(storage.State{})
					Expect(err).NotTo(HaveOccurred())

					_, err = os.Stat(filepath.Join(tempDir, directory))
					Expect(os.IsNotExist(err)).To(Equal(expectToBeDeleted))
				},
				Entry(".bbl", ".bbl", true),
				Entry("terraform", "terraform", true),
				Entry("bosh-deployment", "bosh-deployment", true),
				Entry("jumpbox-deployment", "jumpbox-deployment", true),
				Entry("vars", "vars", true),
				Entry("non-bbl directory", "foo", false),
			)

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

	DescribeTable("get dirs returns the path to an existing directory",
		func(subdirectory string, getDirsFunc func() (string, error)) {
			expectedDir := filepath.Join(tempDir, subdirectory)

			os.MkdirAll(expectedDir, os.ModePerm)

			actualDir, err := getDirsFunc()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualDir).To(Equal(expectedDir))

			os.RemoveAll(expectedDir)
		},
		Entry("cloudconfig", filepath.Join(".bbl", "cloudconfig"), func() (string, error) { return store.GetCloudConfigDir() }),
		Entry("state", "", func() (string, error) { return store.GetStateDir(), nil }),
		Entry("dot-bbl", ".bbl", func() (string, error) { return store.GetBblDir() }),
		Entry("vars", "vars", func() (string, error) { return store.GetVarsDir() }),
		Entry("terraform", "terraform", func() (string, error) { return store.GetTerraformDir() }),
		Entry("bosh-deployment", "bosh-deployment", func() (string, error) { return store.GetDirectorDeploymentDir() }),
		Entry("jumpbox-deployment", "jumpbox-deployment", func() (string, error) { return store.GetJumpboxDeploymentDir() }),
	)

	DescribeTable("get dirs creates a directory that does not already exist",
		func(subdirectory string, getDirsFunc func() (string, error)) {
			expectedDir := filepath.Join(tempDir, subdirectory)

			actualDir, err := getDirsFunc()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualDir).To(Equal(expectedDir))

			_, err = os.Stat(actualDir)
			Expect(err).NotTo(HaveOccurred())

			os.RemoveAll(expectedDir)
		},
		Entry("cloudconfig", filepath.Join(".bbl", "cloudconfig"), func() (string, error) { return store.GetCloudConfigDir() }),
		Entry("dot-bbl", ".bbl", func() (string, error) { return store.GetBblDir() }),
		Entry("vars", "vars", func() (string, error) { return store.GetVarsDir() }),
		Entry("terraform", "terraform", func() (string, error) { return store.GetTerraformDir() }),
		Entry("bosh-deployment", "bosh-deployment", func() (string, error) { return store.GetDirectorDeploymentDir() }),
		Entry("jumpbox-deployment", "jumpbox-deployment", func() (string, error) { return store.GetJumpboxDeploymentDir() }),
	)

	DescribeTable("get dirs returns an error when the subdirectory cannot be created",
		func(subdirectory string, getDirsFunc func() (string, error)) {
			expectedDir := filepath.Join(tempDir, subdirectory)
			_, err := os.Create(expectedDir)
			Expect(err).NotTo(HaveOccurred())

			_, err = getDirsFunc()
			Expect(err).To(MatchError(ContainSubstring("not a directory")))

			os.RemoveAll(expectedDir)
		},
		Entry("dot-bbl", ".bbl", func() (string, error) { return store.GetBblDir() }),
		Entry("vars", "vars", func() (string, error) { return store.GetVarsDir() }),
		Entry("terraform", "terraform", func() (string, error) { return store.GetTerraformDir() }),
		Entry("bosh-deployment", "bosh-deployment", func() (string, error) { return store.GetDirectorDeploymentDir() }),
		Entry("jumpbox-deployment", "jumpbox-deployment", func() (string, error) { return store.GetJumpboxDeploymentDir() }),
	)

	Describe("GetCloudConfigDir", func() {
		var expectedDir string

		BeforeEach(func() {
			expectedDir = filepath.Join(tempDir, ".bbl", "cloudconfig")
		})

		AfterEach(func() {
			os.RemoveAll(expectedDir)
		})

		Context("if the .bbl subdirectory exists", func() {
			BeforeEach(func() {
				os.MkdirAll(filepath.Join(tempDir, ".bbl"), os.ModePerm)
			})

			It("returns the path to the .bbl/cloudconfig directory", func() {
				cloudConfigDir, err := store.GetCloudConfigDir()
				Expect(err).NotTo(HaveOccurred())
				Expect(cloudConfigDir).To(Equal(expectedDir))
			})
		})

		Context("failure cases", func() {
			Context("when the .bbl/cloudconfig subdirectory does not exist and cannot be created", func() {
				BeforeEach(func() {
					os.Mkdir(filepath.Join(tempDir, ".bbl"), os.ModePerm)
					// create a file called .bbl/cloudconfig to cause name collision with the directory to be created
					_, err := os.Create(expectedDir)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					cloudConfigDir, err := store.GetCloudConfigDir()
					Expect(err).To(MatchError(ContainSubstring("not a directory")))
					Expect(cloudConfigDir).To(Equal(""))
				})
			})
		})
	})
})
