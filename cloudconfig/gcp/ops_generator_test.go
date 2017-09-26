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
	"github.com/pivotal-cf-experimental/gomegamatchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCPOpsGenerator", func() {
	Describe("Generate", func() {
		var (
			terraformManager *fakes.TerraformManager
			opsGenerator     gcp.OpsGenerator

			incomingState   storage.State
			expectedOpsFile []byte
		)

		BeforeEach(func() {
			terraformManager = &fakes.TerraformManager{}

			incomingState = storage.State{
				IAAS:    "gcp",
				TFState: "some-tf-state",
				GCP: storage.GCP{
					Region: "us-east1",
					Zones:  []string{"us-east1-b", "us-east1-c", "us-east1-d"},
				},
			}

			terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
				"network_name":       "some-network-name",
				"subnetwork_name":    "some-subnetwork-name",
				"bosh_open_tag_name": "some-bosh-tag",
				"internal_tag_name":  "some-internal-tag",
			}

			var err error
			expectedOpsFile, err = ioutil.ReadFile(filepath.Join("fixtures", "gcp-ops.yml"))
			Expect(err).NotTo(HaveOccurred())

			opsGenerator = gcp.NewOpsGenerator(terraformManager)
		})

		It("returns an ops file to transform base cloud config into gcp specific cloud config", func() {
			opsYAML, err := opsGenerator.Generate(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(incomingState))

			Expect(opsYAML).To(gomegamatchers.MatchYAML(expectedOpsFile))
		})

		DescribeTable("returns an ops file with additional vm extensions to support lb",
			func(lbType string, lbOutputs map[string]interface{}) {
				incomingState.LB.Type = lbType

				expectedLBOpsFile, err := ioutil.ReadFile(filepath.Join("fixtures", fmt.Sprintf("gcp-%s-lb-ops.yml", lbType)))
				Expect(err).NotTo(HaveOccurred())

				expectedOps := strings.Join([]string{string(expectedOpsFile), string(expectedLBOpsFile)}, "\n")

				terraformManager.GetOutputsCall.Returns.Outputs = lbOutputs

				opsYAML, err := opsGenerator.Generate(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(incomingState))

				Expect(opsYAML).To(gomegamatchers.MatchYAML(expectedOps))
			},
			Entry("cf load balancer exists", "cf",
				map[string]interface{}{
					"network_name":           "some-network-name",
					"subnetwork_name":        "some-subnetwork-name",
					"bosh_open_tag_name":     "some-bosh-tag",
					"internal_tag_name":      "some-internal-tag",
					"router_backend_service": "router-backend-service",
					"ws_target_pool":         "ws-target-pool",
					"ssh_proxy_target_pool":  "ssh-proxy-target-pool",
					"tcp_router_target_pool": "tcp-router-target-pool",
				}),
			Entry("concourse load balancer exists", "concourse",
				map[string]interface{}{
					"network_name":          "some-network-name",
					"subnetwork_name":       "some-subnetwork-name",
					"bosh_open_tag_name":    "some-bosh-tag",
					"internal_tag_name":     "some-internal-tag",
					"concourse_target_pool": "concourse-target-pool",
				}),
		)

		Context("failure cases", func() {
			Context("when terraform output provider fails to retrieve", func() {
				BeforeEach(func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("failed to output")
				})

				It("returns an error", func() {
					_, err := opsGenerator.Generate(storage.State{})
					Expect(err).To(MatchError("failed to output"))
				})
			})

			Context("when ops fail to marshal", func() {
				BeforeEach(func() {
					gcp.SetMarshal(func(interface{}) ([]byte, error) {
						return []byte{}, errors.New("failed to marshal")
					})
				})

				It("returns an error", func() {
					_, err := opsGenerator.Generate(storage.State{})
					Expect(err).To(MatchError("failed to marshal"))
					gcp.ResetMarshal()
				})
			})
		})
	})
})
