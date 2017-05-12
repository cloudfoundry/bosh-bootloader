package aws_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform/aws"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OutputGenerator", func() {
	var (
		executor        *fakes.TerraformExecutor
		outputGenerator aws.OutputGenerator
	)

	BeforeEach(func() {
		executor = &fakes.TerraformExecutor{}
		executor.OutputsCall.Returns.Outputs = map[string]interface{}{
			"bosh_eip":                             "some-bosh-eip",
			"bosh_url":                             "some-bosh-url",
			"bosh_user_access_key":                 "some-bosh-user-access-key",
			"bosh_user_secret_access_key":          "some-bosh-user-secret-access_key",
			"bosh_subnet_id":                       "some-bosh-subnet-id",
			"bosh_subnet_availability_zone":        "some-bosh-subnet-availability-zone",
			"bosh_security_group":                  "some-bosh-security-group",
			"internal_security_group":              "some-internal-security-group",
			"internal_subnet_ids":                  "some-internal-subnet-ids",
			"internal_subnet_cidrs":                "some-internal-subnet-cidrs",
			"lb_subnet_ids":                        "some-lb-subnet-ids",
			"lb_subnet_availability_zones":         "some-lb-subnet-availability-zones",
			"lb_subnet_cidrs":                      "some-lb-subnet-cidrs",
			"concourse_lb_name":                    "some-concourse-lb-name",
			"concourse_lb_url":                     "some-concourse-lb-url",
			"concourse_lb_internal_security_group": "some-concourse-internal-security-group",
			"cf_ssh_lb_security_group":             "some-cf-ssh-lb-security_group",
			"cf_ssh_lb_internal_security_group":    "some-cf-ssh-proxy-internal-security-group",
			"cf_router_lb_security_group":          "some-cf-router-lb-security_group",
			"cf_router_lb_internal_security_group": "some-cf-router-internal-security-group",
			"cf_tcp_lb_security_group":             "some-cf-tcp-lb-security_group",
			"cf_tcp_lb_internal_security_group":    "some-cf-tcp-lb-internal-security_group",
			"cf_ssh_lb_name":                       "some-cf-ssh-proxy-lb",
			"cf_ssh_lb_url":                        "some-cf-ssh-proxy-lb-url",
			"cf_router_lb_name":                    "some-cf-router-lb",
			"cf_router_lb_url":                     "some-cf-router-lb-url",
			"cf_tcp_lb_name":                       "some-cf-tcp-lb-name",
			"cf_tcp_lb_url":                        "some-cf-tcp-lb-url",
			"env_dns_zone_name_servers":            []interface{}{"some-name-server-1", "some-name-server-2"},
			"nat_eip":                              "some-nat-eip",
			"vpc_id":                               "some-vpc-id",
		}

		outputGenerator = aws.NewOutputGenerator(executor)
	})

	Context("when no lb exists", func() {
		It("returns all terraform outputs except lb related outputs", func() {
			outputs, err := outputGenerator.Generate(storage.State{
				IAAS:    "aws",
				EnvID:   "some-env-id",
				TFState: "some-tf-state",
				LB: storage.LB{
					Type:   "",
					Domain: "",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(executor.OutputsCall.Receives.TFState).To(Equal("some-tf-state"))

			Expect(outputs).To(Equal(map[string]interface{}{
				"az":                      "some-bosh-subnet-availability-zone",
				"external_ip":             "some-bosh-eip",
				"director_address":        "some-bosh-url",
				"access_key_id":           "some-bosh-user-access-key",
				"secret_access_key":       "some-bosh-user-secret-access_key",
				"subnet_id":               "some-bosh-subnet-id",
				"default_security_groups": "some-bosh-security-group",
				"internal_security_group": "some-internal-security-group",
				"internal_subnet_ids":     "some-internal-subnet-ids",
				"internal_subnet_cidrs":   "some-internal-subnet-cidrs",
				"vpc_id":                  "some-vpc-id",
			}))
		})
	})

	Context("when cf lbs exist", func() {
		It("returns all terraform outputs including cf lb related outputs", func() {
			outputs, err := outputGenerator.Generate(storage.State{
				IAAS:    "aws",
				EnvID:   "some-env-id",
				TFState: "some-tf-state",
				LB: storage.LB{
					Type:   "cf",
					Domain: "some-domain",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(executor.OutputsCall.Receives.TFState).To(Equal("some-tf-state"))

			Expect(outputs).To(Equal(map[string]interface{}{
				"az":                                   "some-bosh-subnet-availability-zone",
				"external_ip":                          "some-bosh-eip",
				"director_address":                     "some-bosh-url",
				"access_key_id":                        "some-bosh-user-access-key",
				"secret_access_key":                    "some-bosh-user-secret-access_key",
				"subnet_id":                            "some-bosh-subnet-id",
				"default_security_groups":              "some-bosh-security-group",
				"internal_security_group":              "some-internal-security-group",
				"internal_subnet_ids":                  "some-internal-subnet-ids",
				"internal_subnet_cidrs":                "some-internal-subnet-cidrs",
				"cf_router_load_balancer":              "some-cf-router-lb",
				"cf_router_load_balancer_url":          "some-cf-router-lb-url",
				"cf_router_internal_security_group":    "some-cf-router-internal-security-group",
				"cf_ssh_proxy_load_balancer":           "some-cf-ssh-proxy-lb",
				"cf_ssh_proxy_load_balancer_url":       "some-cf-ssh-proxy-lb-url",
				"cf_ssh_proxy_internal_security_group": "some-cf-ssh-proxy-internal-security-group",
				"cf_system_domain_dns_servers":         []string{"some-name-server-1", "some-name-server-2"},
				"vpc_id":                               "some-vpc-id",
			}))
		})
	})

	Context("when the concourse lb exists", func() {
		It("returns all terraform outputs including concourse lb related outputs", func() {
			outputs, err := outputGenerator.Generate(storage.State{
				IAAS:    "aws",
				EnvID:   "some-env-id",
				TFState: "some-tf-state",
				LB: storage.LB{
					Type:   "concourse",
					Domain: "",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(executor.OutputsCall.Receives.TFState).To(Equal("some-tf-state"))

			Expect(outputs).To(Equal(map[string]interface{}{
				"az":                                "some-bosh-subnet-availability-zone",
				"external_ip":                       "some-bosh-eip",
				"director_address":                  "some-bosh-url",
				"access_key_id":                     "some-bosh-user-access-key",
				"secret_access_key":                 "some-bosh-user-secret-access_key",
				"subnet_id":                         "some-bosh-subnet-id",
				"default_security_groups":           "some-bosh-security-group",
				"internal_security_group":           "some-internal-security-group",
				"internal_subnet_ids":               "some-internal-subnet-ids",
				"internal_subnet_cidrs":             "some-internal-subnet-cidrs",
				"concourse_load_balancer":           "some-concourse-lb-name",
				"concourse_load_balancer_url":       "some-concourse-lb-url",
				"concourse_internal_security_group": "some-concourse-internal-security-group",
				"vpc_id": "some-vpc-id",
			}))
		})
	})

	Context("failure cases", func() {
		Context("when the executor fails to retrieve the outputs", func() {
			It("returns an error", func() {
				executor.OutputsCall.Returns.Error = errors.New("no can do")

				_, err := outputGenerator.Generate(storage.State{})
				Expect(err).To(MatchError("no can do"))
			})
		})
	})
})
