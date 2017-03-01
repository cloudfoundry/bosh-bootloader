package bosh_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

const (
	variablesYAML = `admin_password: some-admin-password
director_ssl:
  ca: some-ca
  certificate: some-certificate
  private_key: some-private-key
`
)

var _ = Describe("Manager", func() {
	Describe("Create", func() {
		var (
			stackManager            *fakes.StackManager
			boshExecutor            *fakes.BOSHExecutor
			terraformOutputProvider *fakes.TerraformOutputProvider
			boshManager             bosh.Manager
			incomingGCPState        storage.State
			incomingAWSState        storage.State
			variablesMap            map[interface{}]interface{}
		)

		BeforeEach(func() {
			terraformOutputProvider = &fakes.TerraformOutputProvider{}
			stackManager = &fakes.StackManager{}
			boshExecutor = &fakes.BOSHExecutor{}
			boshManager = bosh.NewManager(boshExecutor, terraformOutputProvider, stackManager)

			terraformOutputProvider.GetCall.Returns.Outputs = terraform.Outputs{
				NetworkName:     "some-network",
				SubnetworkName:  "some-subnetwork",
				BOSHTag:         "some-bosh-open-tag",
				InternalTag:     "some-internal-tag",
				ExternalIP:      "some-external-ip",
				DirectorAddress: "some-director-address",
			}

			stackManager.DescribeCall.Returns.Stack = cloudformation.Stack{
				Outputs: map[string]string{
					"BOSHSubnetAZ":            "some-bosh-subnet-az",
					"BOSHUserAccessKey":       "some-bosh-user-access-key",
					"BOSHUserSecretAccessKey": "some-bosh-user-secret-access-key",
					"BOSHSecurityGroup":       "some-bosh-security-group",
					"BOSHSubnet":              "some-bosh-subnet",
					"BOSHEIP":                 "some-bosh-elastic-ip",
					"BOSHURL":                 "some-bosh-url",
				},
			}

			incomingGCPState = storage.State{
				IAAS:  "gcp",
				EnvID: "some-env-id",
				KeyPair: storage.KeyPair{
					PrivateKey: "some-private-key",
				},
				GCP: storage.GCP{
					Zone:              "some-zone",
					ProjectID:         "some-project-id",
					ServiceAccountKey: "some-credential-json",
				},
				BOSH: storage.BOSH{
					State: map[string]interface{}{
						"some-key": "some-value",
					},
				},
				TFState: "some-tf-state",
				LB: storage.LB{
					Type: "cf",
				},
			}

			incomingAWSState = storage.State{
				IAAS:  "aws",
				EnvID: "some-env-id",
				KeyPair: storage.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
				},
				AWS: storage.AWS{
					Region: "some-region",
				},
				Stack: storage.Stack{
					Name: "some-stack",
				},
				BOSH: storage.BOSH{
					State: map[string]interface{}{
						"some-key": "some-value",
					},
				},
				LB: storage.LB{
					Type: "cf",
				},
			}
			variablesMap = map[interface{}]interface{}{
				"admin_password": "some-admin-password",
				"director_ssl": map[interface{}]interface{}{
					"ca":          "some-ca",
					"certificate": "some-certificate",
					"private_key": "some-private-key",
				},
			}

			boshExecutor.InterpolateCall.Returns.Output = bosh.InterpolateOutput{
				Manifest:  "some-manifest",
				Variables: variablesMap,
			}
			boshExecutor.CreateEnvCall.Returns.Output = bosh.CreateEnvOutput{
				State: map[string]interface{}{
					"some-new-key": "some-new-value",
				},
			}
		})

		Context("when iaas is gcp", func() {
			It("queries values from terraform output provider", func() {
				_, err := boshManager.Create(incomingGCPState, []byte{})
				Expect(err).NotTo(HaveOccurred())

				Expect(stackManager.DescribeCall.CallCount).To(Equal(0))
				Expect(terraformOutputProvider.GetCall.Receives.TFState).To(Equal("some-tf-state"))
				Expect(terraformOutputProvider.GetCall.Receives.LBType).To(Equal("cf"))
			})
		})

		Context("when iaas is aws", func() {
			It("queries values from stack", func() {
				_, err := boshManager.Create(incomingAWSState, []byte{})
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformOutputProvider.GetCall.CallCount).To(Equal(0))
				Expect(stackManager.DescribeCall.Receives.StackName).To(Equal("some-stack"))
			})
		})

		DescribeTable("generates a bosh manifest", func(incomingStateFunc func() storage.State,
			expectedInterpolateInput bosh.InterpolateInput) {
			_, err := boshManager.Create(incomingStateFunc(), []byte("some-ops-file"))
			Expect(err).NotTo(HaveOccurred())

			Expect(boshExecutor.InterpolateCall.Receives.InterpolateInput).To(Equal(expectedInterpolateInput))
		},
			Entry("for gcp", func() storage.State {
				return incomingGCPState
			}, bosh.InterpolateInput{
				IAAS: "gcp",
				DeploymentVars: `internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
external_ip: some-external-ip
zone: some-zone
network: some-network
subnetwork: some-subnetwork
tags: [some-bosh-open-tag, some-internal-tag]
project_id: some-project-id
gcp_credentials_json: 'some-credential-json'`,
				BOSHState: map[string]interface{}{
					"some-key": "some-value",
				},
				Variables: "",
				OpsFile:   []byte("some-ops-file"),
			}),
			Entry("for aws", func() storage.State {
				return incomingAWSState
			}, bosh.InterpolateInput{
				IAAS: "aws",
				DeploymentVars: `internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
external_ip: some-bosh-elastic-ip
az: some-bosh-subnet-az
subnet_id: some-bosh-subnet
access_key_id: some-bosh-user-access-key
secret_access_key: some-bosh-user-secret-access-key
default_key_name: some-keypair-name
default_security_groups: [some-bosh-security-group]
region: some-region
private_key: |-
  some-private-key`,
				BOSHState: map[string]interface{}{
					"some-key": "some-value",
				},
				Variables: "",
				OpsFile:   []byte("some-ops-file"),
			}),
		)

		It("creates a bosh environment", func() {
			_, err := boshManager.Create(incomingGCPState, []byte{})
			Expect(err).NotTo(HaveOccurred())

			Expect(boshExecutor.CreateEnvCall.Receives.Input).To(Equal(bosh.CreateEnvInput{
				Manifest: "some-manifest",
				State: map[string]interface{}{
					"some-key": "some-value",
				},
				Variables: variablesYAML,
			}))
		})

		Context("for gcp", func() {
			It("returns a state with a proper bosh state", func() {
				state, err := boshManager.Create(incomingGCPState, []byte{})
				Expect(err).NotTo(HaveOccurred())

				Expect(state).To(Equal(storage.State{
					IAAS:  "gcp",
					EnvID: "some-env-id",
					KeyPair: storage.KeyPair{
						PrivateKey: "some-private-key",
					},
					GCP: storage.GCP{
						Zone:              "some-zone",
						ProjectID:         "some-project-id",
						ServiceAccountKey: "some-credential-json",
					},
					BOSH: storage.BOSH{
						State: map[string]interface{}{
							"some-new-key": "some-new-value",
						},
						Variables:              variablesYAML,
						Manifest:               "some-manifest",
						DirectorName:           "bosh-some-env-id",
						DirectorAddress:        "some-director-address",
						DirectorUsername:       "admin",
						DirectorPassword:       "some-admin-password",
						DirectorSSLCA:          "some-ca",
						DirectorSSLCertificate: "some-certificate",
						DirectorSSLPrivateKey:  "some-private-key",
					},
					TFState: "some-tf-state",
					LB: storage.LB{
						Type: "cf",
					},
				}))
			})
		})

		Context("for aws", func() {
			It("returns a state with a proper bosh state", func() {
				state, err := boshManager.Create(incomingAWSState, []byte{})
				Expect(err).NotTo(HaveOccurred())

				Expect(state).To(Equal(storage.State{
					IAAS:  "aws",
					EnvID: "some-env-id",
					KeyPair: storage.KeyPair{
						Name:       "some-keypair-name",
						PrivateKey: "some-private-key",
					},
					AWS: storage.AWS{
						Region: "some-region",
					},
					Stack: storage.Stack{
						Name: "some-stack",
					},
					BOSH: storage.BOSH{
						State: map[string]interface{}{
							"some-new-key": "some-new-value",
						},
						Variables:              variablesYAML,
						Manifest:               "some-manifest",
						DirectorName:           "bosh-some-env-id",
						DirectorAddress:        "some-bosh-url",
						DirectorUsername:       "admin",
						DirectorPassword:       "some-admin-password",
						DirectorSSLCA:          "some-ca",
						DirectorSSLCertificate: "some-certificate",
						DirectorSSLPrivateKey:  "some-private-key",
					},
					LB: storage.LB{
						Type: "cf",
					},
				}))
			})
		})

		Context("failure cases", func() {
			It("returns an error when terraform output provider fails", func() {
				terraformOutputProvider.GetCall.Returns.Error = errors.New("failed to output")
				_, err := boshManager.Create(storage.State{
					IAAS: "gcp",
				}, []byte{})
				Expect(err).To(MatchError("failed to output"))
			})

			It("returns an error when the stack manager fails", func() {
				stackManager.DescribeCall.Returns.Error = errors.New("failed to get stack")
				_, err := boshManager.Create(storage.State{
					IAAS: "aws",
				}, []byte{})
				Expect(err).To(MatchError("failed to get stack"))
			})

			It("returns an error when an invalid iaas is provided", func() {
				_, err := boshManager.Create(storage.State{}, []byte{})
				Expect(err).To(MatchError("A valid IAAS was not provided"))
			})

			It("returns an error when the executor's interpolate call fails", func() {
				boshExecutor.InterpolateCall.Returns.Error = errors.New("failed to interpolate")
				_, err := boshManager.Create(storage.State{
					IAAS: "gcp",
				}, []byte{})
				Expect(err).To(MatchError("failed to interpolate"))
			})

			It("returns an error when the executor's create env call fails with non create env error", func() {
				boshExecutor.CreateEnvCall.Returns.Error = errors.New("failed to create")
				_, err := boshManager.Create(storage.State{
					IAAS: "gcp",
				}, []byte{})
				Expect(err).To(MatchError("failed to create"))
			})

			Context("when the executor's create env call fails with create env error", func() {
				var (
					incomingState storage.State
					expectedError bosh.ManagerCreateError
					expectedState storage.State
				)

				BeforeEach(func() {
					incomingState = storage.State{
						IAAS: "aws",
					}

					boshState := map[string]interface{}{
						"partial": "bosh-state",
					}
					createEnvError := bosh.NewCreateEnvError(boshState, errors.New("failed to create env"))
					boshExecutor.CreateEnvCall.Returns.Error = createEnvError

					expectedState = incomingState
					expectedState.BOSH = storage.BOSH{
						Manifest:  "some-manifest",
						State:     boshState,
						Variables: variablesYAML,
					}
					expectedError = bosh.NewManagerCreateError(expectedState, createEnvError)
				})

				It("returns a bosh manager create error with a valid state", func() {
					_, err := boshManager.Create(incomingState, []byte{})
					Expect(err).To(MatchError(expectedError))
				})
			})
		})
	})

	Describe("Delete", func() {
		var (
			stackManager            *fakes.StackManager
			boshExecutor            *fakes.BOSHExecutor
			terraformOutputProvider *fakes.TerraformOutputProvider
			boshManager             bosh.Manager
		)

		BeforeEach(func() {
			terraformOutputProvider = &fakes.TerraformOutputProvider{}
			stackManager = &fakes.StackManager{}
			boshExecutor = &fakes.BOSHExecutor{}
			boshManager = bosh.NewManager(boshExecutor, terraformOutputProvider, stackManager)
		})

		It("calls delete env", func() {
			err := boshManager.Delete(storage.State{
				BOSH: storage.BOSH{
					Manifest: "some-manifest",
					State: map[string]interface{}{
						"key": "value",
					},
					Variables: variablesYAML,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(boshExecutor.DeleteEnvCall.Receives.Input).To(Equal(bosh.DeleteEnvInput{
				Manifest: "some-manifest",
				State: map[string]interface{}{
					"key": "value",
				},
				Variables: variablesYAML,
			}))
		})

		Context("failure cases", func() {
			Context("when the executor's delete env call fails with delete env error", func() {
				var (
					incomingState storage.State
					expectedError bosh.ManagerDeleteError
					expectedState storage.State
				)

				BeforeEach(func() {
					incomingState = storage.State{
						IAAS: "aws",
						BOSH: storage.BOSH{
							Manifest: "some-manifest",
							State: map[string]interface{}{
								"key": "value",
							},
							Variables: variablesYAML,
						},
					}

					boshState := map[string]interface{}{
						"partial": "bosh-state",
					}
					deleteEnvError := bosh.NewDeleteEnvError(boshState, errors.New("failed to delete env"))
					boshExecutor.DeleteEnvCall.Returns.Error = deleteEnvError

					expectedState = incomingState
					expectedState.BOSH = storage.BOSH{
						Manifest:  "some-manifest",
						State:     boshState,
						Variables: variablesYAML,
					}
					expectedError = bosh.NewManagerDeleteError(expectedState, deleteEnvError)
				})

				It("returns a bosh manager delete error with a valid state", func() {
					err := boshManager.Delete(incomingState)
					Expect(err).To(MatchError(expectedError))
				})
			})

			It("returns an error when the delete env fails", func() {
				boshExecutor.DeleteEnvCall.Returns.Error = errors.New("failed to delete")
				err := boshManager.Delete(storage.State{})
				Expect(err).To(MatchError("failed to delete"))
			})
		})
	})

	Describe("GetDeploymentVars", func() {
		var (
			stackManager            *fakes.StackManager
			boshExecutor            *fakes.BOSHExecutor
			terraformOutputProvider *fakes.TerraformOutputProvider
			boshManager             bosh.Manager
			incomingGCPState        storage.State
			incomingAWSState        storage.State
		)

		BeforeEach(func() {
			terraformOutputProvider = &fakes.TerraformOutputProvider{}
			stackManager = &fakes.StackManager{}
			boshExecutor = &fakes.BOSHExecutor{}
			boshManager = bosh.NewManager(boshExecutor, terraformOutputProvider, stackManager)

			terraformOutputProvider.GetCall.Returns.Outputs = terraform.Outputs{
				NetworkName:     "some-network",
				SubnetworkName:  "some-subnetwork",
				BOSHTag:         "some-bosh-open-tag",
				InternalTag:     "some-internal-tag",
				ExternalIP:      "some-external-ip",
				DirectorAddress: "some-director-address",
			}

			stackManager.DescribeCall.Returns.Stack = cloudformation.Stack{
				Outputs: map[string]string{
					"BOSHSubnetAZ":            "some-bosh-subnet-az",
					"BOSHUserAccessKey":       "some-bosh-user-access-key",
					"BOSHUserSecretAccessKey": "some-bosh-user-secret-access-key",
					"BOSHSecurityGroup":       "some-bosh-security-group",
					"BOSHSubnet":              "some-bosh-subnet",
					"BOSHEIP":                 "some-bosh-elastic-ip",
					"BOSHURL":                 "some-bosh-url",
				},
			}

			incomingGCPState = storage.State{
				IAAS:  "gcp",
				EnvID: "some-env-id",
				KeyPair: storage.KeyPair{
					PrivateKey: "some-private-key",
				},
				GCP: storage.GCP{
					Zone:              "some-zone",
					ProjectID:         "some-project-id",
					ServiceAccountKey: "some-credential-json",
				},
				BOSH: storage.BOSH{
					State: map[string]interface{}{
						"some-key": "some-value",
					},
				},
				TFState: "some-tf-state",
				LB: storage.LB{
					Type: "cf",
				},
			}

			incomingAWSState = storage.State{
				IAAS:  "aws",
				EnvID: "some-env-id",
				KeyPair: storage.KeyPair{
					Name:       "some-keypair-name",
					PrivateKey: "some-private-key",
				},
				AWS: storage.AWS{
					Region: "some-region",
				},
				Stack: storage.Stack{
					Name: "some-stack",
				},
				BOSH: storage.BOSH{
					State: map[string]interface{}{
						"some-key": "some-value",
					},
				},
				LB: storage.LB{
					Type: "cf",
				},
			}
		})

		Context("gcp", func() {
			It("returns a correct yaml string of bosh deployment variables", func() {
				vars, err := boshManager.GetDeploymentVars(incomingGCPState)
				Expect(err).NotTo(HaveOccurred())
				Expect(vars).To(Equal(`internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
external_ip: some-external-ip
zone: some-zone
network: some-network
subnetwork: some-subnetwork
tags: [some-bosh-open-tag, some-internal-tag]
project_id: some-project-id
gcp_credentials_json: 'some-credential-json'`))
			})
		})

		Context("aws", func() {
			It("returns a correct yaml string of bosh deployment variables", func() {
				vars, err := boshManager.GetDeploymentVars(incomingAWSState)
				Expect(err).NotTo(HaveOccurred())
				Expect(vars).To(Equal(`internal_cidr: 10.0.0.0/24
internal_gw: 10.0.0.1
internal_ip: 10.0.0.6
director_name: bosh-some-env-id
external_ip: some-bosh-elastic-ip
az: some-bosh-subnet-az
subnet_id: some-bosh-subnet
access_key_id: some-bosh-user-access-key
secret_access_key: some-bosh-user-secret-access-key
default_key_name: some-keypair-name
default_security_groups: [some-bosh-security-group]
region: some-region
private_key: |-
  some-private-key`))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the terraform output provider fails", func() {
				terraformOutputProvider.GetCall.Returns.Error = errors.New("failed to output")
				_, err := boshManager.GetDeploymentVars(storage.State{
					IAAS: "gcp",
				})
				Expect(err).To(MatchError("failed to output"))
			})

			It("returns an error when the stack manager fails", func() {
				stackManager.DescribeCall.Returns.Error = errors.New("failed to describe")
				_, err := boshManager.GetDeploymentVars(storage.State{
					IAAS: "aws",
				})
				Expect(err).To(MatchError("failed to describe"))
			})
		})
	})
})
