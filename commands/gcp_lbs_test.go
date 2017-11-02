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

var _ = Describe("GCPLBs", func() {

	var (
		command commands.GCPLBs

		terraformManager *fakes.TerraformManager
		logger           *fakes.Logger

		incomingState storage.State
	)

	BeforeEach(func() {
		terraformManager = &fakes.TerraformManager{}
		terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{Map: map[string]interface{}{
			"router_lb_ip":     "some-router-lb-ip",
			"ssh_proxy_lb_ip":  "some-ssh-proxy-lb-ip",
			"tcp_router_lb_ip": "some-tcp-router-lb-ip",
			"ws_lb_ip":         "some-ws-lb-ip",
			"credhub_lb_ip":    "some-credhub-lb-ip",
			"concourse_lb_ip":  "some-concourse-lb-ip",
		}}
		logger = &fakes.Logger{}

		command = commands.NewGCPLBs(terraformManager, logger)
	})

	Describe("Execute", func() {
		It("prints LB ips for lb type cf", func() {
			incomingState.LB = storage.LB{
				Type: "cf",
			}
			err := command.Execute([]string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
			Expect(logger.PrintfCall.Messages).To(ConsistOf([]string{
				"CF Router LB: some-router-lb-ip\n",
				"CF SSH Proxy LB: some-ssh-proxy-lb-ip\n",
				"CF TCP Router LB: some-tcp-router-lb-ip\n",
				"CF WebSocket LB: some-ws-lb-ip\n",
				"CF Credhub LB: some-credhub-lb-ip\n",
			}))
		})

		Context("when the domain is specified", func() {
			BeforeEach(func() {
				terraformManager.GetOutputsCall.Returns.Outputs = terraform.Outputs{Map: map[string]interface{}{
					"router_lb_ip":              "some-router-lb-ip",
					"ssh_proxy_lb_ip":           "some-ssh-proxy-lb-ip",
					"tcp_router_lb_ip":          "some-tcp-router-lb-ip",
					"ws_lb_ip":                  "some-ws-lb-ip",
					"credhub_lb_ip":             "some-credhub-lb-ip",
					"concourse_lb_ip":           "some-concourse-lb-ip",
					"system_domain_dns_servers": []string{"name-server-1.", "name-server-2."},
				}}
			})

			It("prints LB ips for lb type cf in human readable format", func() {
				incomingState.LB = storage.LB{
					Type:   "cf",
					Domain: "some-domain",
				}
				err := command.Execute([]string{}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
				Expect(logger.PrintfCall.Messages).To(ConsistOf([]string{
					"CF Router LB: some-router-lb-ip\n",
					"CF SSH Proxy LB: some-ssh-proxy-lb-ip\n",
					"CF TCP Router LB: some-tcp-router-lb-ip\n",
					"CF WebSocket LB: some-ws-lb-ip\n",
					"CF Credhub LB: some-credhub-lb-ip\n",
					"CF System Domain DNS servers: name-server-1. name-server-2.\n",
				}))
			})

			Context("when the json flag is provided", func() {
				It("prints LB ips for lb type cf in json format", func() {
					incomingState.LB = storage.LB{
						Type:   "cf",
						Domain: "some-domain",
					}
					err := command.Execute([]string{"--json"}, incomingState)
					Expect(err).NotTo(HaveOccurred())

					Expect(logger.PrintlnCall.Receives.Message).To(MatchJSON(`{
							"cf_router_lb": "some-router-lb-ip",
							"cf_ssh_proxy_lb": "some-ssh-proxy-lb-ip",
							"cf_tcp_router_lb": "some-tcp-router-lb-ip",
							"cf_websocket_lb": "some-ws-lb-ip",
							"cf_credhub_lb": "some-credhub-lb-ip",
							"cf_system_domain_dns_servers": [
								"name-server-1.",
								"name-server-2."
							]
						}`))
				})
			})
		})

		It("prints LB ips for lb type concourse", func() {
			incomingState.LB = storage.LB{
				Type: "concourse",
			}
			err := command.Execute([]string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
			Expect(logger.PrintfCall.Messages).To(ConsistOf([]string{
				"Concourse LB: some-concourse-lb-ip\n",
			}))
		})

		Context("failure cases", func() {
			Context("when terraform output provider fails", func() {
				BeforeEach(func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("failed to return terraform output")
				})

				It("returns an error", func() {
					err := command.Execute([]string{}, incomingState)
					Expect(err).To(MatchError("failed to return terraform output"))
				})
			})

			Context("when no lb type is found", func() {
				BeforeEach(func() {
					incomingState.LB = storage.LB{
						Type: "",
					}
				})

				It("returns an nice error message", func() {
					err := command.Execute([]string{}, incomingState)
					Expect(err).To(MatchError("no lbs found"))
				})
			})
		})
	})
})
