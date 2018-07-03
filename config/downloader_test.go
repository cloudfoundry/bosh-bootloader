package config_test

import (
	"github.com/cloudfoundry/bbl-state-resource/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BBL State Downloader", func() {
	Context("GCP", func() {
		BeforeEach(func() {
			downloader = config.NewDownloader(globalFlags)
		})

		It("makes a storage client with the necessary ids and creds", func() {
			Expect(downloader.fakeCloudStorage.NewObjectCall.Receives.GCPCreds).To(Equal("some-gcp-creds"))
			Expect(downloader.fakeCloudStorage.NewObjectCall.Receives.GCPRegion).To(Equal("some-gcp-region"))
			Expect(downloader.fakeCloudStorage.NewObjectCall.Receives.GCPProject).To(Equal("some-gcp-project"))
			Expect(downloader.fakeCloudStorage.NewObjectCall.Receives.Bucket).To(Equal("some-gcp-bucket"))
			Expect(downloader.fakeCloudStorage.NewObjectCall.Receives.EnvID).To(Equal("some-gcp-project"))
		})

		Context("when we can read a tarball directly out of GCP", func() {
			var fakeObject *fakes.Object
			BeforeEach(func() {
				downloader.fakeCloudStorage.NewObjectCall.Returns = fakeObject
				fakeObject.NewReaderCall.Returns.ReadCloser = fakeReadCloser // probably needs to be a tarball?
			})

			It("untars the bbl state from the reader", func() {
			})
		})

		Context("when creds are invalid", func() {
			It("errors clearly", func() {
			})
		})
	})
})
