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
		It("receives state and returns a map of terraform variables", func() {
			inputs, err := inputGenerator.Generate(storage.State{
				EnvID: "banana",
				OpenStack: storage.OpenStack{
					InternalCidr:         "10.0.0.0/16",
					ExternalIP:           "external-ip",
					AuthURL:              "auth-url",
					AZ:                   "az",
					DefaultKeyName:       "key-name",
					DefaultSecurityGroup: "security-group",
					NetworkID:            "network-id",
					Password:             "password",
					Username:             "username",
					Project:              "project",
					Domain:               "domain",
					Region:               "region",
					PrivateKey:           "private-key",
				},
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(inputs).To(Equal(map[string]interface{}{
				"env_id":                 "banana",
				"internal_cidr":          "10.0.0.0/16",
				"internal_gw":            "10.0.0.1",
				"director_ip":            "10.0.0.6",
				"external_ip":            "external-ip",
				"auth_url":               "auth-url",
				"az":                     "az",
				"default_key_name":       "key-name",
				"default_security_group": "security-group",
				"net_id":                 "network-id",
				"openstack_project":      "project",
				"openstack_domain":       "domain",
				"region":                 "region",
				"private_key":            "private-key",
			}))
		})
	})
	Describe("Credentials", func() {
		It("returns the vsphere credentials", func() {
			state := storage.State{
				OpenStack: storage.OpenStack{
					Username: "the-user",
					Password: "the-password",
				},
			}

			credentials := inputGenerator.Credentials(state)

			Expect(credentials).To(Equal(map[string]string{
				"openstack_username": "the-user",
				"openstack_password": "the-password",
			}))
		})
	})
})
