package gcp_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/cloudconfig/gcp"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	"github.com/pivotal-cf-experimental/gomegamatchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCPOpsGenerator", func() {
	Describe("Generate", func() {
		var (
			zones            *fakes.Zones
			terraformManager *fakes.TerraformManager
			opsGenerator     gcp.OpsGenerator

			incomingState   storage.State
			expectedOpsFile []byte
		)

		BeforeEach(func() {
			terraformManager = &fakes.TerraformManager{}
			zones = &fakes.Zones{}

			incomingState = storage.State{
				IAAS:    "gcp",
				TFState: "some-tf-state",
				GCP: storage.GCP{
					Region: "us-east1",
				},
			}

			zones.GetCall.Returns.Zones = []string{"us-east1-b", "us-east1-c", "us-east1-d"}
			terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{
				NetworkName:    "some-network-name",
				SubnetworkName: "some-subnetwork-name",
				BOSHTag:        "some-bosh-tag",
				InternalTag:    "some-internal-tag",
			}

			var err error
			expectedOpsFile, err = ioutil.ReadFile(filepath.Join("fixtures", "gcp-ops.yml"))
			Expect(err).NotTo(HaveOccurred())

			opsGenerator = gcp.NewOpsGenerator(terraformManager, zones)
		})

		It("returns an ops file to transform base cloud config into gcp specific cloud config", func() {
			opsYAML, err := opsGenerator.Generate(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(zones.GetCall.Receives.Region).To(Equal("us-east1"))
			Expect(terraformManager.GetOutputsCall.Receives.TFState).To(Equal("some-tf-state"))

			Expect(opsYAML).To(gomegamatchers.MatchYAML(expectedOpsFile))
		})

		DescribeTable("returns an ops file with additional vm extensions to support lb", func(lbType string, lbOutputs terraform.Outputs) {
			incomingState.LB.Type = lbType

			expectedLBOpsFile, err := ioutil.ReadFile(filepath.Join("fixtures", fmt.Sprintf("gcp-%s-lb-ops.yml", lbType)))
			Expect(err).NotTo(HaveOccurred())

			expectedOps := strings.Join([]string{string(expectedOpsFile), string(expectedLBOpsFile)}, "\n")

			terraformManager.GetOutputsCall.Returns.Outputs = lbOutputs

			opsYAML, err := opsGenerator.Generate(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformManager.GetOutputsCall.Receives.LBType).To(Equal(lbType))

			Expect(opsYAML).To(gomegamatchers.MatchYAML(expectedOps))
		},
			Entry("cf load balancer exists", "cf",
				terraform.Outputs{
					NetworkName:          "some-network-name",
					SubnetworkName:       "some-subnetwork-name",
					BOSHTag:              "some-bosh-tag",
					InternalTag:          "some-internal-tag",
					RouterBackendService: "router-backend-service",
					WSTargetPool:         "ws-target-pool",
					SSHProxyTargetPool:   "ssh-proxy-target-pool",
					TCPRouterTargetPool:  "tcp-router-target-pool",
				}),
			Entry("concourse load balancer exists", "concourse",
				terraform.Outputs{
					NetworkName:         "some-network-name",
					SubnetworkName:      "some-subnetwork-name",
					BOSHTag:             "some-bosh-tag",
					InternalTag:         "some-internal-tag",
					ConcourseTargetPool: "concourse-target-pool",
				}),
		)

		Context("failure cases", func() {
			It("returns an error when terraform output provider fails to retrieve", func() {
				terraformManager.GetOutputsCall.Returns.Error = errors.New("failed to output")
				_, err := opsGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("failed to output"))
			})

			It("returns an error when it fails to parse a cidr block", func() {
				zones.GetCall.Returns.Zones = []string{"z", "z", "z", "z", "z", "z", "z", "z", "z", "z", "z", "z", "z", "z", "z", "z", "z", "z", "z", "z"}
				_, err := opsGenerator.Generate(storage.State{})
				Expect(err).To(MatchError(`invalid ip, 10.0.256.0 has values out of range`))
			})

			It("returns an error when ops fail to marshal", func() {
				gcp.SetMarshal(func(interface{}) ([]byte, error) {
					return []byte{}, errors.New("failed to marshal")
				})
				_, err := opsGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("failed to marshal"))
				gcp.ResetMarshal()
			})
		})
	})
})
