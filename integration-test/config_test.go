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

			config, err := integration.LoadConfig(configurationFilePath)
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

				_, err := integration.LoadConfig(configurationFilePath)
				Expect(err).To(MatchError("aws access key id is missing"))
			})

			It("returns an error if aws access key id is missing", func() {
				configurationFilePath := writeConfigurationFile(`{
					"AWSAccessKeyID": "some-aws-access-key-id",
					"AWSRegion": "some-region"
				}`)

				_, err := integration.LoadConfig(configurationFilePath)
				Expect(err).To(MatchError("aws secret access key is missing"))
			})

			It("returns an error if aws access key id is missing", func() {
				configurationFilePath := writeConfigurationFile(`{
					"AWSAccessKeyID": "some-aws-access-key-id",
					"AWSSecretAccessKey": "some-aws-secret-access-key"
				}`)

				_, err := integration.LoadConfig(configurationFilePath)
				Expect(err).To(MatchError("aws region is missing"))
			})

			It("returns an error if it cannot open the config file", func() {
				_, err := integration.LoadConfig("nonexistent-file")
				Expect(err).To(MatchError("open nonexistent-file: no such file or directory"))
			})

			It("returns an error if it cannot parse json in config file", func() {
				configurationFilePath := writeConfigurationFile(`%%%%`)

				_, err := integration.LoadConfig(configurationFilePath)
				Expect(err).To(MatchError("invalid character '%' looking for beginning of value"))
			})
		})
	})

	Describe("ConfigPath", func() {
		var configPath string

		BeforeEach(func() {
			configPath = os.Getenv("BIT_CONFIG")
		})

		AfterEach(func() {
			os.Setenv("BIT_CONFIG", configPath)
		})

		Context("when a valid path is set", func() {
			It("returns the path", func() {
				os.Setenv("BIT_CONFIG", "/tmp/some-config.json")
				path, err := integration.ConfigPath()
				Expect(err).NotTo(HaveOccurred())
				Expect(path).To(Equal("/tmp/some-config.json"))
			})
		})

		Context("when path is not set", func() {
			It("returns an error", func() {
				os.Setenv("BIT_CONFIG", "")
				_, err := integration.ConfigPath()
				Expect(err).To(MatchError(`$BIT_CONFIG "" does not specify an absolute path to test config file`))
			})
		})

		Context("when the path is not absolute", func() {
			It("returns an error", func() {
				os.Setenv("BIT_CONFIG", "some/path.json")
				_, err := integration.ConfigPath()
				Expect(err).To(MatchError(`$BIT_CONFIG "some/path.json" does not specify an absolute path to test config file`))
			})
		})
	})
})
