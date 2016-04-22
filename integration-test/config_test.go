package integration_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("config", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("LoadConfig", func() {
		var configPath string

		BeforeEach(func() {
			configPath = os.Getenv("BIT_CONFIG")
		})

		AfterEach(func() {
			os.Setenv("BIT_CONFIG", configPath)
		})

		var writeConfigurationFile = func(json string) string {
			configurationFilePath := filepath.Join(tempDir, "config.json")

			err := ioutil.WriteFile(configurationFilePath, []byte(json), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			return configurationFilePath
		}

		It("returns a valid config from a path", func() {
			configurationFilePath := writeConfigurationFile(`{
				"AWSAccessKeyID": "some-aws-access-key-id",
				"AWSSecretAccessKey": "some-aws-secret-access-key",
				"AWSRegion": "some-region"
			}`)

			os.Setenv("BIT_CONFIG", configurationFilePath)

			config, err := integration.LoadConfig()
			Expect(err).NotTo(HaveOccurred())

			Expect(config).To(Equal(integration.Config{
				AWSAccessKeyID:     "some-aws-access-key-id",
				AWSSecretAccessKey: "some-aws-secret-access-key",
				AWSRegion:          "some-region",
			}))
		})

		Context("failure cases", func() {
			It("returns an error if aws access key id is missing", func() {
				configurationFilePath := writeConfigurationFile(`{
					"AWSSecretAccessKey": "some-aws-secret-access-key",
					"AWSRegion": "some-region"
				}`)

				os.Setenv("BIT_CONFIG", configurationFilePath)
				_, err := integration.LoadConfig()
				Expect(err).To(MatchError("aws access key id is missing"))
			})

			It("returns an error if aws access key id is missing", func() {
				configurationFilePath := writeConfigurationFile(`{
					"AWSAccessKeyID": "some-aws-access-key-id",
					"AWSRegion": "some-region"
				}`)

				os.Setenv("BIT_CONFIG", configurationFilePath)
				_, err := integration.LoadConfig()
				Expect(err).To(MatchError("aws secret access key is missing"))
			})

			It("returns an error if aws access key id is missing", func() {
				configurationFilePath := writeConfigurationFile(`{
					"AWSAccessKeyID": "some-aws-access-key-id",
					"AWSSecretAccessKey": "some-aws-secret-access-key"
				}`)

				os.Setenv("BIT_CONFIG", configurationFilePath)
				_, err := integration.LoadConfig()
				Expect(err).To(MatchError("aws region is missing"))
			})

			It("returns an error if it cannot open the config file", func() {
				os.Setenv("BIT_CONFIG", "/nonexistent-file")
				_, err := integration.LoadConfig()
				Expect(err).To(MatchError("open /nonexistent-file: no such file or directory"))
			})

			It("returns an error if it cannot parse json in config file", func() {
				configurationFilePath := writeConfigurationFile(`%%%%`)
				os.Setenv("BIT_CONFIG", configurationFilePath)

				_, err := integration.LoadConfig()
				Expect(err).To(MatchError("invalid character '%' looking for beginning of value"))
			})

			It("returns an error when the path is not set", func() {
				os.Setenv("BIT_CONFIG", "")
				_, err := integration.LoadConfig()
				Expect(err).To(MatchError(`$BIT_CONFIG "" does not specify an absolute path to test config file`))
			})

			It("returns an error when the path is not absolute", func() {
				os.Setenv("BIT_CONFIG", "some/path.json")
				_, err := integration.LoadConfig()
				Expect(err).To(MatchError(`$BIT_CONFIG "some/path.json" does not specify an absolute path to test config file`))
			})
		})
	})
})
