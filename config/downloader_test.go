package config_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/araddon/gou"
	"github.com/cloudfoundry/bosh-bootloader/config"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/lytics/cloudstorage"
	"github.com/lytics/cloudstorage/awss3"
	"github.com/lytics/cloudstorage/google"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BBL State Downloader", func() {
	var (
		downloader  config.Downloader
		backend     *fakes.Backend
		globalFlags config.GlobalFlags
	)

	BeforeEach(func() {
		backend = &fakes.Backend{}
		downloader = config.NewDownloader(backend)

		backend.GetStateCall.Returns.State = ioutil.NopCloser(bytes.NewReader([]byte("some bbl state file contents")))
	})

	Context("AWS", func() {
		var awsAuthSettings gou.JsonHelper
		BeforeEach(func() {
			globalFlags = config.GlobalFlags{
				IAAS:               "aws",
				AWSAccessKeyID:     "some-access-key-id",
				AWSSecretAccessKey: "some-secret-access-key",
				AWSRegion:          "some-region",
				StateBucket:        "some-aws-bbl-states",
				EnvID:              "some-aws-env",
			}

			awsAuthSettings = make(gou.JsonHelper)
			awsAuthSettings[awss3.ConfKeyAccessKey] = "some-access-key-id"
			awsAuthSettings[awss3.ConfKeyAccessSecret] = "some-secret-access-key"
		})

		It("uses an aws config to call the backend", func() {
			file, err := downloader.Download(globalFlags)
			Expect(err).NotTo(HaveOccurred())

			Expect(backend.GetStateCall.CallCount).To(Equal(1))
			Expect(backend.GetStateCall.Receives.Name).To(Equal("some-aws-env"))
			Expect(backend.GetStateCall.Receives.Config).To(Equal(
				cloudstorage.Config{
					Type:       awss3.StoreType,
					AuthMethod: awss3.AuthAccessKey,
					Bucket:     "some-aws-bbl-states",
					Settings:   awsAuthSettings,
					Region:     "some-region",
				},
			))

			Expect(file).NotTo(BeNil())
			contents, err := ioutil.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(Equal([]byte("some bbl state file contents")))
		})
	})

	Context("GCP", func() {
		var serviceAccountKey cloudstorage.JwtConf
		BeforeEach(func() {
			serviceAccountKey = cloudstorage.JwtConf{
				PrivateKeyID: "some-private-key-id",
				PrivateKey:   "some-private-key",
				ClientEmail:  "some-client-email",
				ClientID:     "some-client-id",
				Type:         "some-type",
				ProjectID:    "some-gcp-project",
			}

			marshalledServiceAccountKey, err := json.Marshal(serviceAccountKey)
			Expect(err).NotTo(HaveOccurred())

			globalFlags = config.GlobalFlags{
				IAAS:                 "gcp",
				GCPServiceAccountKey: string(marshalledServiceAccountKey),
				GCPRegion:            "some-region",
				StateBucket:          "some-gcp-project-bbl-states",
				EnvID:                "some-gcp-env",
			}

		})

		It("uses a gcp config to call the backend", func() {
			file, err := downloader.Download(globalFlags)
			Expect(err).NotTo(HaveOccurred())

			Expect(backend.GetStateCall.CallCount).To(Equal(1))
			Expect(backend.GetStateCall.Receives.Name).To(Equal("some-gcp-env"))
			Expect(backend.GetStateCall.Receives.Config).To(Equal(
				cloudstorage.Config{
					Type:       google.StoreType,
					AuthMethod: google.AuthJWTKeySource,
					Project:    "some-gcp-project",
					Bucket:     "some-gcp-project-bbl-states",
					JwtConf:    &serviceAccountKey,
					// TmpDir:     "/tmp/localcache/google",
				},
			))

			Expect(file).NotTo(BeNil())
			contents, err := ioutil.ReadAll(file)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(Equal([]byte("some bbl state file contents")))
		})

		Context("when creds are invalid", func() {
			BeforeEach(func() {
				globalFlags.GCPServiceAccountKey = "not a real service account key"
			})

			It("returns an error", func() {
				_, err := downloader.Download(globalFlags)
				Expect(err).To(MatchError("invalid GCP service account key"))
				Expect(backend.GetStateCall.CallCount).To(Equal(0))
			})
		})

		Context("when the backend returns an error", func() {
			BeforeEach(func() {
				backend.GetStateCall.Returns.Error = errors.New("jackfruit")
			})

			It("returns an error", func() {
				_, err := downloader.Download(globalFlags)
				Expect(err).To(MatchError("jackfruit"))
			})
		})
	})
})
