package bosh_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deployments", func() {
	It("retrieves all deployments from the director", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.URL.Path).To(Equal("/deployments"))
			Expect(r.Method).To(Equal("GET"))

			username, password, ok := r.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(username).To(Equal("some-username"))
			Expect(password).To(Equal("some-password"))

			w.Write([]byte(`[
				{
					"name": "cardeployment",
					"releases": [{"name": "sports-release", "version": "ferrari-version"}],
					"stemcells": [{"name": "sedan-stemcell", "version": "lexus-version"}],
					"cloud_config": "some-cloud-config"
				},
				{
					"name": "animaldeployment",
					"releases": [{"name": "ungulate-release", "version": "cow-version"}, {"name": "ungulate-release", "version": "yak-version"}],
					"stemcells": [{"name": "canid-stemcell", "version": "wolf-version"}, {"name": "canid-stemcell", "version": "dog-version"}],
					"cloud_config": "some-other-cloud-config"
				}
			]`))
		}))

		client := bosh.NewClient(bosh.Config{
			URL:      server.URL,
			Username: "some-username",
			Password: "some-password",
		})

		deployments, err := client.Deployments()
		Expect(err).NotTo(HaveOccurred())
		Expect(deployments).To(Equal(
			[]bosh.Deployment{
				{
					Name: "cardeployment",
					Releases: []bosh.Release{
						{
							Name:     "sports-release",
							Versions: []string{"ferrari-version"},
						},
					},
					Stemcells: []bosh.Stemcell{
						{
							Name:     "sedan-stemcell",
							Versions: []string{"lexus-version"},
						},
					},
					CloudConfig: "some-cloud-config",
				},
				{
					Name: "animaldeployment",
					Releases: []bosh.Release{
						{
							Name:     "ungulate-release",
							Versions: []string{"cow-version", "yak-version"},
						},
					},
					Stemcells: []bosh.Stemcell{{
						Name:     "canid-stemcell",
						Versions: []string{"wolf-version", "dog-version"},
					}},
					CloudConfig: "some-other-cloud-config",
				},
			},
		))
	})

	Context("failure cases", func() {
		It("error on a malformed URL", func() {
			client := bosh.NewClient(bosh.Config{
				URL:      "%%%%%%%%",
				Username: "some-username",
				Password: "some-password",
			})

			_, err := client.Deployments()
			Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
		})

		It("error on an empty URL", func() {
			client := bosh.NewClient(bosh.Config{
				URL:      "",
				Username: "some-username",
				Password: "some-password",
			})

			_, err := client.Deployments()
			Expect(err).To(MatchError(ContainSubstring("unsupported protocol")))
		})

		It("errors on an unexpected status code with a body", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				w.Write([]byte("More Info"))
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			_, err := client.Deployments()
			Expect(err).To(MatchError("unexpected response 502 Bad Gateway:\nMore Info"))
		})

		It("error on malformed JSON", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`%%%%%%%%`))
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			_, err := client.Deployments()
			Expect(err).To(MatchError(ContainSubstring("invalid character")))
		})

		It("returns an error on a bogus response body", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				w.Write([]byte("More Info"))
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			bosh.SetBodyReader(func(io.Reader) ([]byte, error) {
				return nil, errors.New("a bad read happened")
			})

			_, err := client.Deployments()

			Expect(err).To(MatchError("a bad read happened"))
		})
	})
})
