package gcp_test

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/cloudconfig/gcp"
	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("CloudConfigGenerator", func() {
	Describe("Generate", func() {
		var (
			cloudConfigGenerator gcp.CloudConfigGenerator
		)

		BeforeEach(func() {
			cloudConfigGenerator = gcp.NewCloudConfigGenerator()
		})

		AfterEach(func() {
			gcp.ResetUnmarshal()
		})

		It("generates a cloud config with no load balancers", func() {
			cloudConfig, err := cloudConfigGenerator.Generate(gcp.CloudConfigInput{
				AZs:            []string{"us-east1-b", "us-east1-c", "us-east1-d"},
				Tags:           []string{"some-tag"},
				NetworkName:    "some-network-name",
				SubnetworkName: "some-subnetwork-name",
			})
			Expect(err).NotTo(HaveOccurred())

			buf, err := ioutil.ReadFile("fixtures/cloud-config-no-lb.yml")
			Expect(err).NotTo(HaveOccurred())

			output, err := yaml.Marshal(cloudConfig)
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(MatchYAML(string(buf)))
		})

		It("generates a cloud config with a concourse load balancer", func() {
			cloudConfig, err := cloudConfigGenerator.Generate(gcp.CloudConfigInput{
				AZs:                 []string{"us-east1-b", "us-east1-c", "us-east1-d"},
				Tags:                []string{"some-tag"},
				NetworkName:         "some-network-name",
				SubnetworkName:      "some-subnetwork-name",
				ConcourseTargetPool: "concourse-target-pool",
			})
			Expect(err).NotTo(HaveOccurred())

			buf, err := ioutil.ReadFile("fixtures/cloud-config-concourse-lb.yml")
			Expect(err).NotTo(HaveOccurred())

			output, err := yaml.Marshal(cloudConfig)
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(MatchYAML(string(buf)))
		})

		It("generates a cloud config with a cf load balancers", func() {
			cloudConfig, err := cloudConfigGenerator.Generate(gcp.CloudConfigInput{
				AZs:            []string{"us-east1-b", "us-east1-c", "us-east1-d"},
				Tags:           []string{"some-tag"},
				NetworkName:    "some-network-name",
				SubnetworkName: "some-subnetwork-name",
				CFBackends: gcp.CFBackends{
					Router:    "router-backend-service",
					SSHProxy:  "ssh-proxy-target-pool",
					TCPRouter: "tcp-router-target-pool",
					WS:        "ws-target-pool",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			buf, err := ioutil.ReadFile("fixtures/cloud-config-cf-lb.yml")
			Expect(err).NotTo(HaveOccurred())

			output, err := yaml.Marshal(cloudConfig)
			Expect(err).NotTo(HaveOccurred())

			Expect(output).To(MatchYAML(string(buf)))
		})

		Context("failure cases", func() {
			It("returns an error when the base cloud config template fails to marshal", func() {
				gcp.SetUnmarshal(func([]byte, interface{}) error {
					return errors.New("failed to unmarshal")
				})

				_, err := cloudConfigGenerator.Generate(gcp.CloudConfigInput{
					AZs:            []string{"us-east1-b", "us-east1-c", "us-east1-d"},
					Tags:           []string{"some-tag", "some-other-tag"},
					NetworkName:    "some-network-name",
					SubnetworkName: "some-subnetwork-name",
				})

				Expect(err).To(MatchError("failed to unmarshal"))

			})

			It("returns an error when it fails to generate networks for manifest", func() {
				azs := []string{}
				for i := 0; i < 255; i++ {
					azs = append(azs, fmt.Sprintf("az%d", i))
				}

				_, err := cloudConfigGenerator.Generate(gcp.CloudConfigInput{
					AZs:            azs,
					Tags:           []string{"some-tag", "some-other-tag"},
					NetworkName:    "some-network-name",
					SubnetworkName: "some-subnetwork-name",
				})

				Expect(err).To(MatchError(ContainSubstring("invalid ip")))
			})
		})
	})
})
