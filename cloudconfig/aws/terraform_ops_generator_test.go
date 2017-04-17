package aws_test

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/cloudconfig/aws"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/pivotal-cf-experimental/gomegamatchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TerraformOpsGenerator", func() {
	Describe("Generate", func() {
		var (
			availabilityZoneRetriever *fakes.AvailabilityZoneRetriever
			terraformManager          *fakes.TerraformManager
			opsGenerator              aws.TerraformOpsGenerator

			incomingState   storage.State
			expectedOpsFile []byte
		)

		BeforeEach(func() {
			availabilityZoneRetriever = &fakes.AvailabilityZoneRetriever{}
			terraformManager = &fakes.TerraformManager{}

			incomingState = storage.State{
				IAAS: "aws",
				AWS: storage.AWS{
					Region: "us-east-1",
				},
				TFState: "some-tf-state",
			}

			availabilityZoneRetriever.RetrieveCall.Returns.AZs = []string{
				"us-east-1a",
				"us-east-1b",
				"us-east-1c",
			}

			terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
				"internal_subnet_cidrs": []interface{}{
					"10.0.16.0/20",
					"10.0.32.0/20",
					"10.0.48.0/20",
				},
				"internal_subnet_ids": []interface{}{
					"some-subnet-1",
					"some-subnet-2",
					"some-subnet-3",
				},
				"internal_security_group": "some-internal-security-group",
			}

			var err error
			expectedOpsFile, err = ioutil.ReadFile(filepath.Join("fixtures", "aws-ops.yml"))
			Expect(err).NotTo(HaveOccurred())

			opsGenerator = aws.NewTerraformOpsGenerator(availabilityZoneRetriever, terraformManager)
		})

		It("returns an ops file to transform base cloud config into aws specific cloud config", func() {
			opsYAML, err := opsGenerator.Generate(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(availabilityZoneRetriever.RetrieveCall.Receives.Region).To(Equal("us-east-1"))
			Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(incomingState))

			Expect(opsYAML).To(gomegamatchers.MatchYAML(expectedOpsFile))
		})

		Context("failure cases", func() {
			It("returns an error when az retriever fails to retrieve", func() {
				availabilityZoneRetriever.RetrieveCall.Returns.Error = errors.New("failed to retrieve")
				_, err := opsGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("failed to retrieve"))
			})

			It("returns an error when the infrastructure manager fails to describe stack", func() {
				terraformManager.GetOutputsCall.Returns.Error = errors.New("failed to get outputs")
				_, err := opsGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("failed to get outputs"))
			})

			It("returns an error when it fails to parse a cidr block", func() {
				terraformManager.GetOutputsCall.Returns.Outputs["internal_subnet_cidrs"] = []interface{}{
					"****",
				}
				_, err := opsGenerator.Generate(storage.State{})
				Expect(err).To(MatchError(`"****" cannot parse CIDR block`))
			})

			It("returns an error when ops fails to marshal", func() {
				aws.SetMarshal(func(interface{}) ([]byte, error) {
					return []byte{}, errors.New("failed to marshal")
				})
				_, err := opsGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("failed to marshal"))
				aws.ResetMarshal()
			})

			It("returns an error when no internal_subnet_cidrs output exists", func() {
				terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
					"internal_subnet_ids": []interface{}{
						"some-subnet-1",
						"some-subnet-2",
						"some-subnet-3",
					},
					"internal_security_group": "some-internal-security-group",
				}

				_, err := opsGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("missing internal subnet cidrs terraform output"))
			})

			It("returns an error when no internal_subnet_ids output exists", func() {
				terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
					"internal_subnet_cidrs": []interface{}{
						"10.0.16.0/20",
						"10.0.32.0/20",
						"10.0.48.0/20",
					},
					"internal_security_group": "some-internal-security-group",
				}

				_, err := opsGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("missing internal subnet ids terraform output"))
			})

			It("returns an error when no internal_security_group output exists", func() {
				terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
					"internal_subnet_cidrs": []interface{}{
						"10.0.16.0/20",
						"10.0.32.0/20",
						"10.0.48.0/20",
					},
					"internal_subnet_ids": []interface{}{
						"some-subnet-1",
						"some-subnet-2",
						"some-subnet-3",
					},
				}

				_, err := opsGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("missing internal security group terraform output"))
			})
		})
	})
})
