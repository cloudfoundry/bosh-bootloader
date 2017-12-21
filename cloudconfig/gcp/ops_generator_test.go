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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCPOpsGenerator", func() {
	var (
		terraformManager *fakes.TerraformManager
		opsGenerator     gcp.OpsGenerator

		incomingState storage.State
	)

	BeforeEach(func() {
		terraformManager = &fakes.TerraformManager{}

		incomingState = storage.State{
			IAAS: "gcp",
			GCP: storage.GCP{
				Region: "us-east1",
				Zones:  []string{"us-east1-b", "us-east1-c", "us-east1-d"},
			},
		}

		terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{Map: map[string]interface{}{
			"network_name":           "some-network-name",
			"subnetwork_name":        "some-subnetwork-name",
			"bosh_open_tag_name":     "some-bosh-tag",
			"internal_tag_name":      "some-internal-tag",
			"router_backend_service": "some-backend-service",
			"ws_target_pool":         "some-ws-target-pool",
			"ssh_proxy_target_pool":  "some-proxy-target-pool",
			"tcp_router_target_pool": "some-tcp-router-target-pool",
			"concourse_target_pool":  "some-concourse-target-pool",
			"credhub_target_pool":    "some-credhub-target-pool",
		}}

		opsGenerator = gcp.NewOpsGenerator(terraformManager)
	})

	Describe("GenerateVars", func() {
		It("returns the contents of a vars file", func() {
			varsYAML, err := opsGenerator.GenerateVars(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
			Expect(varsYAML).To(MatchYAML(`
bosh_open_tag_name: some-bosh-tag
network_name: some-network-name
subnetwork_name: some-subnetwork-name
internal_tag_name: some-internal-tag
internal_tag_name: some-internal-tag
router_backend_service: some-backend-service
ws_target_pool: some-ws-target-pool
ssh_proxy_target_pool: some-proxy-target-pool
tcp_router_target_pool: some-tcp-router-target-pool
concourse_target_pool: some-concourse-target-pool
credhub_target_pool: some-credhub-target-pool
`))
		})
		Context("when terraform output provider fails to retrieve", func() {
			BeforeEach(func() {
				terraformManager.GetOutputsCall.Returns.Error = errors.New("tomato")
			})

			It("returns an error", func() {
				_, err := opsGenerator.GenerateVars(storage.State{})
				Expect(err).To(MatchError("Get terraform outputs: tomato"))
			})
		})
	})

	Describe("Generate", func() {
		var expectedOpsFile []byte

		BeforeEach(func() {
			var err error
			expectedOpsFile, err = ioutil.ReadFile(filepath.Join("fixtures", "gcp-ops.yml"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an ops file to transform base cloud config into gcp specific cloud config", func() {
			opsYAML, err := opsGenerator.Generate(incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(opsYAML).To(MatchYAML(expectedOpsFile))
		})

		DescribeTable("returns an ops file with additional vm extensions to support lb",
			func(lbType string) {
				incomingState.LB.Type = lbType
				expectedLBOpsFile, err := ioutil.ReadFile(filepath.Join("fixtures", fmt.Sprintf("gcp-%s-lb-ops.yml", lbType)))
				Expect(err).NotTo(HaveOccurred())
				expectedOps := strings.Join([]string{string(expectedOpsFile), string(expectedLBOpsFile)}, "\n")

				opsYAML, err := opsGenerator.Generate(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(opsYAML).To(MatchYAML(expectedOps))
			},
			Entry("cf load balancer exists", "cf"),
			Entry("concourse load balancer exists", "concourse"),
		)

		Context("failure cases", func() {
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
