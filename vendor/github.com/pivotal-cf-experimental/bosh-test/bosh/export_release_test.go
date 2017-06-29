package bosh_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ExportRelease", func() {
	var server *httptest.Server

	BeforeEach(func() {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/releases/export":
				Expect(req.Method).To(Equal("POST"))
				username, password, ok := req.BasicAuth()
				Expect(ok).To(BeTrue())
				Expect(username).To(Equal("some-username"))
				Expect(password).To(Equal("some-password"))

				Expect(req.Header.Get("Content-Type")).To(Equal("application/json"))

				var release map[string]interface{}
				err := json.NewDecoder(req.Body).Decode(&release)
				Expect(err).NotTo(HaveOccurred())
				Expect(release).To(Equal(map[string]interface{}{
					"deployment_name":  "some-deployment-name",
					"release_name":     "some-release-name",
					"release_version":  "some-release-version",
					"stemcell_os":      "some-stemcell-name",
					"stemcell_version": "some-stemcell-version",
				}))

				w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/5", req.Host))
				w.WriteHeader(http.StatusFound)

			case "/tasks/5":
				Expect(req.Method).To(Equal("GET"))
				username, password, ok := req.BasicAuth()
				Expect(ok).To(BeTrue())
				Expect(username).To(Equal("some-username"))
				Expect(password).To(Equal("some-password"))

				w.Write([]byte(`{"id": 5, "state": "done"}`))

			case "/tasks/5/output":
				Expect(req.Method).To(Equal("GET"))
				Expect(req.URL.RawQuery).To(Equal("type=result"))
				username, password, ok := req.BasicAuth()
				Expect(ok).To(BeTrue())
				Expect(username).To(Equal("some-username"))
				Expect(password).To(Equal("some-password"))

				w.Write([]byte(`{"blobstore_id": "some-resource-id"}`))

			default:
				Fail(fmt.Sprintf("unhandled response %s", req.URL.Path))
			}

		}))
	})

	It("exports the release and returns a resource ID for the release artifact", func() {
		client := bosh.NewClient(bosh.Config{
			URL:      server.URL,
			Username: "some-username",
			Password: "some-password",
		})

		resourceID, err := client.ExportRelease("some-deployment-name",
			"some-release-name", "some-release-version",
			"some-stemcell-name", "some-stemcell-version")
		Expect(err).NotTo(HaveOccurred())
		Expect(resourceID).To(Equal("some-resource-id"))
	})

	Context("failure cases", func() {
		Context("when the reqest cannot be created", func() {
			It("returns an error", func() {
				client := bosh.NewClient(bosh.Config{
					URL: "%%%%%",
				})

				_, err := client.ExportRelease("some-deployment-name",
					"some-release-name", "some-release-version",
					"some-stemcell-name", "some-stemcell-version")
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})

		Context("when the request cannot be made", func() {
			It("returns an error", func() {
				client := bosh.NewClient(bosh.Config{
					URL: "",
				})

				_, err := client.ExportRelease("some-deployment-name",
					"some-release-name", "some-release-version",
					"some-stemcell-name", "some-stemcell-version")
				Expect(err).To(MatchError(ContainSubstring("unsupported protocol scheme")))
			})
		})

		Context("when the task status request fails", func() {
			It("returns an error", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Header().Set("Location", "%%%%")
					w.WriteHeader(http.StatusFound)
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.ExportRelease("some-deployment-name",
					"some-release-name", "some-release-version",
					"some-stemcell-name", "some-stemcell-version")
				Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
			})
		})

		Context("when the task result request fails", func() {
			It("returns an error", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					switch req.URL.Path {
					case "/releases/export":
						w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/5", req.Host))
						w.WriteHeader(http.StatusFound)

					case "/tasks/5":
						w.Write([]byte(`{"id": 5, "state": "done"}`))

					case "/tasks/5/output":
						w.WriteHeader(http.StatusTeapot)
						w.Write([]byte("More Info"))

					default:
						Fail(fmt.Sprintf("unhandled response %s", req.URL.Path))
					}
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.ExportRelease("some-deployment-name",
					"some-release-name", "some-release-version",
					"some-stemcell-name", "some-stemcell-version")
				Expect(err).To(MatchError("unexpected response 418 I'm a teapot:\nMore Info"))
			})
		})

		Context("when the task result does not include a blobstore_id", func() {
			It("returns an error", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					switch req.URL.Path {
					case "/releases/export":
						w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/5", req.Host))
						w.WriteHeader(http.StatusFound)

					case "/tasks/5":
						w.Write([]byte(`{"id": 5, "state": "done"}`))

					case "/tasks/5/output":
						w.Write([]byte("{}"))

					default:
						Fail(fmt.Sprintf("unhandled response %s", req.URL.Path))
					}
				}))

				client := bosh.NewClient(bosh.Config{
					URL:      server.URL,
					Username: "some-username",
					Password: "some-password",
				})

				_, err := client.ExportRelease("some-deployment-name",
					"some-release-name", "some-release-version",
					"some-stemcell-name", "some-stemcell-version")
				Expect(err).To(MatchError("could not find \"blobstore_id\" key in task result"))
			})
		})
	})
})
