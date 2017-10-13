package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AWSLBs", func() {
	var (
		command commands.AWSLBs

		terraformManager *fakes.TerraformManager
		logger           *fakes.Logger

		incomingState storage.State
	)

	BeforeEach(func() {
		terraformManager = &fakes.TerraformManager{}
		logger = &fakes.Logger{}

		command = commands.NewAWSLBs(terraformManager, logger)
	})

	Describe("Execute", func() {
		Context("when the lb type is cf", func() {
			BeforeEach(func() {
				incomingState = storage.State{
					IAAS:    "aws",
					TFState: "some-tf-state",
					LB: storage.LB{
						Type: "cf",
					},
				}
				terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{Map: map[string]interface{}{
					"cf_router_lb_name": "some-router-lb-name",
					"cf_router_lb_url":  "some-router-lb-url",
					"cf_ssh_lb_name":    "some-ssh-lb-name",
					"cf_ssh_lb_url":     "some-ssh-lb-url",
					"cf_tcp_lb_name":    "some-tcp-lb-name",
					"cf_tcp_lb_url":     "some-tcp-lb-url",
				}}
			})

			It("prints LB names and URLs for router and ssh proxy", func() {
				err := command.Execute([]string{}, incomingState)

				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PrintfCall.Messages).To(ConsistOf([]string{
					"CF Router LB: some-router-lb-name [some-router-lb-url]\n",
					"CF SSH Proxy LB: some-ssh-lb-name [some-ssh-lb-url]\n",
					"CF TCP Router LB: some-tcp-lb-name [some-tcp-lb-url]\n",
				}))
			})

			Context("when the domain is specified", func() {
				BeforeEach(func() {
					incomingState.LB.Domain = "some-domain"

					terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{Map: map[string]interface{}{
						"cf_router_lb_name":         "some-router-lb-name",
						"cf_router_lb_url":          "some-router-lb-url",
						"cf_ssh_lb_name":            "some-ssh-lb-name",
						"cf_ssh_lb_url":             "some-ssh-lb-url",
						"cf_tcp_lb_name":            "some-tcp-lb-name",
						"cf_tcp_lb_url":             "some-tcp-lb-url",
						"env_dns_zone_name_servers": []string{"name-server-1.", "name-server-2."},
					}}
				})

				It("prints LB names, URLs, and DNS servers", func() {
					err := command.Execute([]string{}, incomingState)

					Expect(err).NotTo(HaveOccurred())

					Expect(logger.PrintfCall.Messages).To(ConsistOf([]string{
						"CF Router LB: some-router-lb-name [some-router-lb-url]\n",
						"CF SSH Proxy LB: some-ssh-lb-name [some-ssh-lb-url]\n",
						"CF TCP Router LB: some-tcp-lb-name [some-tcp-lb-url]\n",
						"CF System Domain DNS servers: name-server-1. name-server-2.\n",
					}))
				})

				Context("when the json flag is provided", func() {
					It("prints LB names, URLs, and DNS servers in json format", func() {
						incomingState.LB = storage.LB{
							Type:   "cf",
							Domain: "some-domain",
						}
						err := command.Execute([]string{"--json"}, incomingState)
						Expect(err).NotTo(HaveOccurred())

						Expect(logger.PrintlnCall.Receives.Message).To(MatchJSON(`{
								"cf_router_lb": "some-router-lb-name",
								"cf_router_lb_url": "some-router-lb-url",
								"cf_ssh_proxy_lb": "some-ssh-lb-name",
								"cf_ssh_proxy_lb_url": "some-ssh-lb-url",
								"cf_tcp_lb": "some-tcp-lb-name",
								"cf_tcp_lb_url":  "some-tcp-lb-url",
								"env_dns_zone_name_servers": [
									"name-server-1.",
									"name-server-2."
								]
							}`))
					})
				})
			})
		})

		Context("when the lb type is concourse", func() {
			BeforeEach(func() {
				incomingState = storage.State{
					IAAS:    "aws",
					TFState: "some-tf-state",
					LB: storage.LB{
						Type: "concourse",
					},
				}
				terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{Map: map[string]interface{}{
					"concourse_lb_name": "some-concourse-lb-name",
					"concourse_lb_url":  "some-concourse-lb-url",
				}}
			})

			It("prints LB name and URL", func() {
				err := command.Execute([]string{}, incomingState)

				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PrintfCall.Messages).To(ConsistOf([]string{
					"Concourse LB: some-concourse-lb-name [some-concourse-lb-url]\n",
				}))
			})
		})

		Context("when lb type is not cf or concourse", func() {
			BeforeEach(func() {
				incomingState = storage.State{
					IAAS:    "aws",
					TFState: "some-tf-state",
					LB: storage.LB{
						Type: "other",
					},
				}
			})

			It("returns error", func() {
				err := command.Execute([]string{}, incomingState)

				Expect(err).To(MatchError("no lbs found"))
			})
		})

		Context("failure cases", func() {
			BeforeEach(func() {
				incomingState = storage.State{
					TFState: "some-tf-state",
				}
			})

			Context("when terraform manager fails", func() {
				It("returns an error", func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("terraform manager failed")

					err := command.Execute([]string{}, incomingState)

					Expect(err).To(MatchError("terraform manager failed"))
				})
			})
		})
	})
})
