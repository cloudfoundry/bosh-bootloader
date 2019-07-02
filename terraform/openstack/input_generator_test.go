package openstack_test

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform/openstack"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InputGenerator", func() {
	var (
		inputGenerator openstack.InputGenerator
	)

	Describe("Generate", func() {
		var state storage.State
		BeforeEach(func() {
			state = storage.State{
				EnvID: "banana",
				OpenStack: storage.OpenStack{
					AuthURL:     "auth-url",
					AZ:          "az",
					NetworkID:   "network-id",
					NetworkName: "network-name",
					Password:    "password",
					Username:    "username",
					Project:     "project",
					Domain:      "domain",
					Region:      "region",
				},
			}
		})

		Context("when optional OpenStack variables have zero values", func() {
			It("receives state and returns a map of required terraform variables", func() {
				inputs, err := inputGenerator.Generate(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(inputs).To(Equal(map[string]interface{}{
					"env_id":            "banana",
					"auth_url":          "auth-url",
					"availability_zone": "az",
					"ext_net_id":        "network-id",
					"ext_net_name":      "network-name",
					"domain_name":       "domain",
					"region_name":       "region",
					"tenant_name":       "project",
				}))
			})
		})

		Context("when optional OpenStack variables are set", func() {
			JustBeforeEach(func() {
				state.OpenStack.CACertFile = "path/to/file"
				state.OpenStack.Insecure = "true"
				state.OpenStack.DNSNameServers = []string{"8.8.8.8", "9.9.9.9"}
			})
			It("receives state and returns a map of required and optional terraform variables", func() {
				inputs, err := inputGenerator.Generate(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(inputs).To(Equal(map[string]interface{}{
					"env_id":            "banana",
					"auth_url":          "auth-url",
					"availability_zone": "az",
					"ext_net_id":        "network-id",
					"ext_net_name":      "network-name",
					"domain_name":       "domain",
					"region_name":       "region",
					"tenant_name":       "project",
					"cacert_file":       "path/to/file",
					"insecure":          "true",
					"dns_nameservers":   []string{"8.8.8.8", "9.9.9.9"},
				}))
			})
		})
	})
	Describe("Credentials", func() {
		It("returns the OpenStack credentials", func() {
			state := storage.State{
				OpenStack: storage.OpenStack{
					Username: "the-user",
					Password: "the-password",
				},
			}

			credentials := inputGenerator.Credentials(state)

			Expect(credentials).To(Equal(map[string]string{
				"user_name": "the-user",
				"password":  "the-password",
			}))
		})
	})
})
