package gcp_test

import (
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform/gcp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InputGenerator", func() {
	var (
		inputGenerator gcp.InputGenerator
		state          storage.State
	)

	BeforeEach(func() {
		var err error
		Expect(err).NotTo(HaveOccurred())

		state = storage.State{
			IAAS:  "gcp",
			EnvID: "some-env-id",
			GCP: storage.GCP{
				ServiceAccountKey:     "some-service-account-key",
				ServiceAccountKeyPath: "/some/service/account/key",
				ProjectID:             "some-project-id",
				Zone:                  "some-zone",
				Zones:                 []string{"zone-1", "zone-2"},
				Region:                "some-region",
			},
			TFState: "some-tf-state",
			LB: storage.LB{
				Type:   "cf",
				Domain: "some-domain",
			},
		}

		inputGenerator = gcp.NewInputGenerator()
	})

	Describe("Generate", func() {
		It("receives BBL state and returns a map of terraform variables", func() {
			inputs, err := inputGenerator.Generate(state)
			Expect(err).NotTo(HaveOccurred())

			Expect(inputs).To(Equal(map[string]interface{}{
				"env_id":        state.EnvID,
				"project_id":    state.GCP.ProjectID,
				"region":        state.GCP.Region,
				"zone":          state.GCP.Zone,
				"zones":         state.GCP.Zones,
				"system_domain": state.LB.Domain,
			}))
		})

		Context("when cert and key are provided", func() {
			BeforeEach(func() {
				state.LB.Cert = "some-cert"
				state.LB.Key = "some-key"
			})

			It("returns a map containing cert and key variables", func() {
				inputs, err := inputGenerator.Generate(state)
				Expect(err).NotTo(HaveOccurred())

				Expect(inputs).To(Equal(map[string]interface{}{
					"env_id":                      state.EnvID,
					"project_id":                  state.GCP.ProjectID,
					"region":                      state.GCP.Region,
					"zone":                        state.GCP.Zone,
					"zones":                       state.GCP.Zones,
					"ssl_certificate":             state.LB.Cert,
					"ssl_certificate_private_key": state.LB.Key,
					"system_domain":               state.LB.Domain,
				}))
			})
		})
	})

	Describe("Credentials", func() {
		It("returns the service account key", func() {
			credentials := inputGenerator.Credentials(state)

			Expect(credentials).To(Equal(map[string]string{
				"credentials": "/some/service/account/key",
			}))
		})
	})
})
