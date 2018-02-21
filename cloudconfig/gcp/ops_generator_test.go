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
			"internal_cidr":          "10.0.0.0/20",
			"subnet_cidr_1":          "10.0.16.0/24",
			"subnet_cidr_2":          "10.0.32.0/24",
			"subnet_cidr_3":          "10.0.48.0/24",
			"network_name":           "some-network-name",
			"subnetwork_name":        "some-subnetwork-name",
			"bosh_open_tag_name":     "some-bosh-tag",
			"internal_tag_name":      "some-internal-tag",
			"router_backend_service": "some-backend-service",
			"ws_target_pool":         "some-ws-target-pool",
			"ssh_proxy_target_pool":  "some-proxy-target-pool",
			"tcp_router_target_pool": "some-tcp-router-target-pool",
			"concourse_target_pool":  "some-concourse-target-pool",
		}}

		opsGenerator = gcp.NewOpsGenerator(terraformManager)
	})

	Describe("GenerateVars", func() {
		It("returns the contents of a vars file", func() {
			varsYAML, err := opsGenerator.GenerateVars(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
			Expect(varsYAML).To(MatchYAML(`
az1_gateway: 10.0.16.1
az1_name: us-east1-b
az1_range: 10.0.16.0/24
az1_reserved_1: 10.0.16.2-10.0.16.3
az1_reserved_2: 10.0.16.255
az1_static: 10.0.16.190-10.0.16.254
az2_gateway: 10.0.32.1
az2_name: us-east1-c
az2_range: 10.0.32.0/24
az2_reserved_1: 10.0.32.2-10.0.32.3
az2_reserved_2: 10.0.32.255
az2_static: 10.0.32.190-10.0.32.254
az3_gateway: 10.0.48.1
az3_name: us-east1-d
az3_range: 10.0.48.0/24
az3_reserved_1: 10.0.48.2-10.0.48.3
az3_reserved_2: 10.0.48.255
az3_static: 10.0.48.190-10.0.48.254
bosh_open_tag_name: some-bosh-tag
network_name: some-network-name
subnetwork_name: some-subnetwork-name
internal_cidr: 10.0.0.0/20
internal_tag_name: some-internal-tag
internal_tag_name: some-internal-tag
router_backend_service: some-backend-service
ws_target_pool: some-ws-target-pool
ssh_proxy_target_pool: some-proxy-target-pool
subnet_cidr_1: 10.0.16.0/24
subnet_cidr_2: 10.0.32.0/24
subnet_cidr_3: 10.0.48.0/24
tcp_router_target_pool: some-tcp-router-target-pool
concourse_target_pool: some-concourse-target-pool
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
