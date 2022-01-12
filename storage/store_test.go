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
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Store", func() {
	var (
		fileIO           *fakes.FileIO
		garbageCollector *fakes.GarbageCollector
		store            storage.Store
		tempDir          string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "")

		fileIO = &fakes.FileIO{}
		garbageCollector = &fakes.GarbageCollector{}

		store = storage.NewStore(tempDir, fileIO, garbageCollector)
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
					BBLVersion: "5.3.0",
					IAAS:       "aws",
					AWS: storage.AWS{
						AccessKeyID:     "some-aws-access-key-id",
						SecretAccessKey: "some-aws-secret-access-key",
						Region:          "some-region",
					},
					Azure: storage.Azure{
						ClientID:       "client-id",
						ClientSecret:   "client-secret",
						Region:         "some-azure-region",
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
					VSphere: storage.VSphere{
						VCenterUser:     "user",
						VCenterPassword: "password",
						VCenterIP:       "ip",
						VCenterDC:       "dc",
						VCenterCluster:  "cluster",
						VCenterRP:       "rp",
						Network:         "network",
						VCenterDS:       "ds",
						SubnetCIDR:      "10.0.0.0/24",
					},
					OpenStack: storage.OpenStack{
						AuthURL:     "auth-url",
						AZ:          "az",
						NetworkID:   "network-id",
						NetworkName: "network-name",
						Password:    "password",
						Username:    "username",
						Project:     "project",
						Domain:      "domain",
						Region:      "region",
					},
					LB: storage.LB{
						Type:   "some-type",
						Cert:   "some-cert",
						Key:    "some-key",
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
						Variables: "some-vars",
						Manifest:  "name: bosh",
					},
					EnvID:   "some-env-id",
					TFState: "some-tf-state",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(filepath.Join(tempDir, "bbl-state.json")))
				Expect(fileIO.WriteFileCall.Receives[0].Mode).To(Equal(os.FileMode(0644)))
				Expect(fileIO.WriteFileCall.Receives[0].Contents).To(MatchJSON(`{
				"version": 14,
				"bblVersion": "5.3.0",
				"iaas": "aws",
				"id": "01020304-0506-0708-0910-111213141516",
				"envID": "some-env-id",
				"aws": {
					"region": "some-region"
				},
				"azure": {
					"region": "some-azure-region"
				},
				"gcp": {
					"zone": "some-zone",
					"region": "some-region",
					"zones": ["some-zone", "some-other-zone"]
				},
				"vsphere": {},
				"openstack": {},
				"cloudstack": {},
				"lb": {
					"type": "some-type",
					"cert": "some-cert",
					"key": "some-key",
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
					"state": {
						"key": "value"
					}
				},
				"tfState": "some-tf-state",
				"latestTFOutput": ""
		    	}`))
			})
		})

		Context("when the state is empty", func() {
			It("calls the garbage collector", func() {
				err := store.Set(storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(garbageCollector.RemoveCall.CallCount).To(Equal(1))
				Expect(garbageCollector.RemoveCall.Receives.Directory).To(Equal(tempDir))
			})

			Context("when the garbage collector fails to clean up", func() {
				BeforeEach(func() {
					garbageCollector.RemoveCall.Returns.Error = errors.New("banana")
				})
				It("returns the error", func() {
					err := store.Set(storage.State{})
					Expect(err).To(MatchError("Garbage collector clean up: banana"))
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

			Context("when the directory does not exist", func() {
				BeforeEach(func() {
					fileIO.StatCall.Returns.Error = errors.New("no such file or directory")
				})

				It("returns an error", func() {
					store = storage.NewStore("non-valid-dir", fileIO, garbageCollector)
					err := store.Set(storage.State{})
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			Context("when it fails to open the bbl-state.json file", func() {
				BeforeEach(func() {
					fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{Error: errors.New("permission denied")}}
				})

				It("returns an error", func() {
					err := store.Set(storage.State{EnvID: "something"})
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})

	DescribeTable("get dirs returns the path to an existing directory",
		func(subdirectory string, getDirsFunc func() (string, error)) {
			expectedDir := filepath.Join(tempDir, subdirectory)

			actualDir, err := getDirsFunc()
			Expect(err).NotTo(HaveOccurred())
			Expect(actualDir).To(Equal(expectedDir))

			if len(subdirectory) > 0 {
				Expect(fileIO.MkdirAllCall.Receives.Dir).To(Equal(expectedDir))
			}
		},
		Entry("cloud-config", "cloud-config", func() (string, error) { return store.GetCloudConfigDir() }),
		Entry("state", "", func() (string, error) { return store.GetStateDir(), nil }),
		Entry("vars", "vars", func() (string, error) { return store.GetVarsDir() }),
		Entry("terraform", "terraform", func() (string, error) { return store.GetTerraformDir() }),
		Entry("bosh-deployment", "bosh-deployment", func() (string, error) { return store.GetDirectorDeploymentDir() }),
		Entry("jumpbox-deployment", "jumpbox-deployment", func() (string, error) { return store.GetJumpboxDeploymentDir() }),
	)

	DescribeTable("get dirs returns an error when the subdirectory cannot be created",
		func(subdirectory string, getDirsFunc func() (string, error)) {
			expectedDir := filepath.Join(tempDir, subdirectory)
			fileIO.MkdirAllCall.Returns.Error = errors.New("not a directory")

			_, err := getDirsFunc()
			Expect(err).To(MatchError(ContainSubstring("not a directory")))
			Expect(fileIO.MkdirAllCall.Receives.Dir).To(Equal(expectedDir))
		},
		Entry("cloud-config", "cloud-config", func() (string, error) { return store.GetCloudConfigDir() }),
		Entry("vars", "vars", func() (string, error) { return store.GetVarsDir() }),
		Entry("terraform", "terraform", func() (string, error) { return store.GetTerraformDir() }),
		Entry("bosh-deployment", "bosh-deployment", func() (string, error) { return store.GetDirectorDeploymentDir() }),
		Entry("jumpbox-deployment", "jumpbox-deployment", func() (string, error) { return store.GetJumpboxDeploymentDir() }),
	)

	Describe("GetCloudConfigDir", func() {
		var expectedCloudConfigPath string

		BeforeEach(func() {
			expectedCloudConfigPath = filepath.Join(tempDir, "cloud-config")
		})

		Context("if the cloud-config subdirectory exists", func() {
			It("returns the path to the cloud-config directory", func() {
				cloudConfigDir, err := store.GetCloudConfigDir()
				Expect(err).NotTo(HaveOccurred())
				Expect(cloudConfigDir).To(Equal(expectedCloudConfigPath))
			})
		})

		Context("failure cases", func() {
			Context("when there is a name collision with an existing file", func() {
				BeforeEach(func() {
					fileIO.MkdirAllCall.Returns.Error = errors.New("not a directory")
				})

				It("returns an error", func() {
					cloudConfigDir, err := store.GetCloudConfigDir()
					Expect(err).To(MatchError("Get cloud-config dir: not a directory"))
					Expect(cloudConfigDir).To(Equal(""))
				})
			})
		})
	})

	Describe("GetRuntimeConfigDir", func() {
		var expectedRuntimeConfigPath string

		BeforeEach(func() {
			expectedRuntimeConfigPath = filepath.Join(tempDir, "runtime-config")
		})

		Context("if the runtime-config subdirectory exists", func() {
			It("returns the path to the runtime-config directory", func() {
				runtimeConfigDir, err := store.GetRuntimeConfigDir()
				Expect(err).NotTo(HaveOccurred())
				Expect(runtimeConfigDir).To(Equal(expectedRuntimeConfigPath))
			})
		})

		Context("failure cases", func() {
			Context("when there is a name collision with an existing file", func() {
				BeforeEach(func() {
					fileIO.MkdirAllCall.Returns.Error = errors.New("not a directory")
				})

				It("returns an error", func() {
					runtimeConfigDir, err := store.GetRuntimeConfigDir()
					Expect(err).To(MatchError("Get runtime-config dir: not a directory"))
					Expect(runtimeConfigDir).To(Equal(""))
				})
			})
		})
	})

	Describe("GetVarsDir", func() {
		Context("when the vars dir is requested but may not exist", func() {
			It("a path is request and may be created and set with restrained permissions", func() {
				_, err := store.GetVarsDir()
				Expect(err).NotTo(HaveOccurred())
				Expect(fileIO.MkdirAllCall.Receives.Perm).To(Equal(os.FileMode(storage.StateMode)))
			})
		})

	})
})
