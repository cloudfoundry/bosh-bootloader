package gcp_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform/gcp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InputGenerator", func() {
	var (
		inputGenerator gcp.InputGenerator

		tempDir string
		state   storage.State
	)

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		gcp.SetTempDir(func(dir, prefix string) (string, error) {
			return tempDir, nil
		})

		state = storage.State{
			IAAS:  "gcp",
			EnvID: "some-env-id",
			GCP: storage.GCP{
				ServiceAccountKey: "some-service-account-key",
				ProjectID:         "some-project-id",
				Zone:              "some-zone",
				Region:            "some-region",
			},
			TFState: "some-tf-state",
			LB: storage.LB{
				Type:   "cf",
				Domain: "some-domain",
			},
		}

		inputGenerator = gcp.NewInputGenerator()
	})

	AfterEach(func() {
		gcp.ResetTempDir()
		gcp.ResetWriteFile()
	})

	It("receives BBL state and returns a map of terraform variables", func() {
		inputs, err := inputGenerator.Generate(state)
		Expect(err).NotTo(HaveOccurred())

		Expect(inputs).To(Equal(map[string]interface{}{
			"env_id":        state.EnvID,
			"project_id":    state.GCP.ProjectID,
			"region":        state.GCP.Region,
			"zone":          state.GCP.Zone,
			"credentials":   filepath.Join(tempDir, "credentials.json"),
			"system_domain": state.LB.Domain,
		}))

		credentials, err := ioutil.ReadFile(inputs["credentials"].(string))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(credentials)).To(Equal("some-service-account-key"))
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
				"credentials":                 filepath.Join(tempDir, "credentials.json"),
				"ssl_certificate":             filepath.Join(tempDir, "cert"),
				"ssl_certificate_private_key": filepath.Join(tempDir, "key"),
				"system_domain":               state.LB.Domain,
			}))

			sslCertificate, err := ioutil.ReadFile(inputs["ssl_certificate"].(string))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(sslCertificate)).To(Equal("some-cert"))

			sslCertificatePrivateKey, err := ioutil.ReadFile(inputs["ssl_certificate_private_key"].(string))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(sslCertificatePrivateKey)).To(Equal("some-key"))
		})
	})

	Context("failure cases", func() {
		It("returns an error if temp dir cannot be created", func() {
			gcp.SetTempDir(func(dir, prefix string) (string, error) {
				return "", errors.New("failed to create temp dir")
			})
			_, err := inputGenerator.Generate(state)
			Expect(err).To(MatchError("failed to create temp dir"))
		})

		It("returns an error if the credentials cannot be written", func() {
			gcp.SetWriteFile(func(filename string, data []byte, perm os.FileMode) error {
				if strings.Contains(filename, "credentials.json") {
					return errors.New("failed to write file")
				}
				return nil
			})
			_, err := inputGenerator.Generate(state)
			Expect(err).To(MatchError("failed to write file"))
		})

		Context("when cert and key are provided", func() {
			BeforeEach(func() {
				state.LB.Cert = "some-cert"
				state.LB.Key = "some-cert"
			})

			It("returns an error if the cert cannot be written", func() {
				gcp.SetWriteFile(func(filename string, data []byte, perm os.FileMode) error {
					if strings.Contains(filename, "cert") {
						return errors.New("failed to write file")
					}
					return nil
				})
				_, err := inputGenerator.Generate(state)
				Expect(err).To(MatchError("failed to write file"))
			})

			It("returns an error if the key cannot be written", func() {
				gcp.SetWriteFile(func(filename string, data []byte, perm os.FileMode) error {
					if strings.Contains(filename, "key") {
						return errors.New("failed to write file")
					}
					return nil
				})
				_, err := inputGenerator.Generate(state)
				Expect(err).To(MatchError("failed to write file"))
			})
		})
	})
})
