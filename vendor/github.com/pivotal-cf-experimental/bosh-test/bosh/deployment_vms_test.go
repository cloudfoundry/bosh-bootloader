package bosh_test

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeploymentVMs", func() {
	It("retrieves the list of deployment VMs given a deployment name", func() {
		var taskCallCount int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(username).To(Equal("some-username"))
			Expect(password).To(Equal("some-password"))

			switch r.URL.Path {
			case "/deployments/some-deployment-name/vms":
				Expect(r.URL.RawQuery).To(Equal("format=full"))
				host, _, err := net.SplitHostPort(r.Host)
				Expect(err).NotTo(HaveOccurred())

				location := &url.URL{
					Scheme: "http",
					Host:   host,
					Path:   "/tasks/1",
				}

				w.Header().Set("Location", location.String())
				w.WriteHeader(http.StatusFound)
			case "/tasks/1":
				w.WriteHeader(http.StatusAccepted)
				w.Write([]byte(`{"state":"done"}`))
				taskCallCount++
			case "/tasks/1/output":
				Expect(r.URL.RawQuery).To(Equal("type=result"))
				Expect(taskCallCount).NotTo(Equal(0))

				w.Write([]byte(`
					{"id": "id-c0", "index": 0, "job_name": "consul_z1", "job_state":"some-state", "ips": ["1.2.3.4"]}
					{"id": "id-e0", "index": 0, "job_name": "etcd_z1", "job_state":"some-state", "ips": ["1.2.3.5"]}
					{"id": "id-e1", "index": 1, "job_name": "etcd_z1", "job_state":"some-other-state", "ips": ["1.2.3.6"]}
					{"id": "id-e2", "index": 2, "job_name": "etcd_z1", "job_state":"some-more-state", "ips": ["1.2.3.7"]}
				`))
			default:
				Fail("unknown route")
			}
		}))

		client := bosh.NewClient(bosh.Config{
			URL:      server.URL,
			Username: "some-username",
			Password: "some-password",
		})

		vms, err := client.DeploymentVMs("some-deployment-name")
		Expect(err).NotTo(HaveOccurred())
		Expect(vms).To(ConsistOf([]bosh.VM{
			{
				ID:      "id-c0",
				Index:   0,
				JobName: "consul_z1",
				State:   "some-state",
				IPs:     []string{"1.2.3.4"},
			},
			{
				ID:      "id-e0",
				Index:   0,
				JobName: "etcd_z1",
				State:   "some-state",
				IPs:     []string{"1.2.3.5"},
			},
			{
				ID:      "id-e1",
				Index:   1,
				JobName: "etcd_z1",
				State:   "some-other-state",
				IPs:     []string{"1.2.3.6"},
			},
			{
				ID:      "id-e2",
				Index:   2,
				JobName: "etcd_z1",
				State:   "some-more-state",
				IPs:     []string{"1.2.3.7"},
			},
		}))
	})

	Context("failure cases", func() {
		It("errors when the URL is malformed", func() {
			client := bosh.NewClient(bosh.Config{
				URL: "http://%%%%%",
			})

			_, err := client.DeploymentVMs("some-deployment-name")
			Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
		})

		It("errors when the protocol scheme is invalid", func() {
			client := bosh.NewClient(bosh.Config{
				URL: "banana://example.com",
			})

			_, err := client.DeploymentVMs("some-deployment-name")
			Expect(err).To(MatchError(ContainSubstring("unsupported protocol")))
		})

		It("errors when checking the task fails", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/deployments/some-deployment-name/vms":
					w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/1", r.Host))
					w.WriteHeader(http.StatusFound)
				case "/tasks/1":
					w.Write([]byte("%%%"))
				default:
					Fail("unexpected route")
				}
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			_, err := client.DeploymentVMs("some-deployment-name")
			Expect(err).To(MatchError(ContainSubstring("invalid character")))
		})

		It("should error on a non StatusFound status code with a body", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.URL.Path).To(Equal("/deployments/some-deployment-name/vms"))
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("More Info"))
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			_, err := client.DeploymentVMs("some-deployment-name")
			Expect(err).To(MatchError("unexpected response 404 Not Found:\nMore Info"))
		})

		It("errors when the redirect URL is malformed", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.URL.Path).To(Equal("/deployments/some-deployment-name/vms"))
				w.Header().Set("Location", "http://%%%%%/tasks/1")
				w.WriteHeader(http.StatusFound)
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			_, err := client.DeploymentVMs("some-deployment-name")
			Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
		})

		It("should error on malformed JSON", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/1", r.Host))
				w.WriteHeader(http.StatusFound)
				w.Write([]byte("%%%%%%\n%%%%%%%%%%%\n"))
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			_, err := client.DeploymentVMs("some-deployment-name")
			Expect(err).To(MatchError(ContainSubstring("invalid character")))
		})

		It("should error on a bogus response body", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/deployments/some-deployment-name/vms":
					w.Header().Set("Location", fmt.Sprintf("http://%s/tasks/1", r.Host))
					w.WriteHeader(http.StatusFound)
				case "/tasks/1":
					w.Write([]byte(`{"state": "done"}`))
				case "/tasks/1/output":
					w.Write([]byte(""))
				default:
					Fail("unexpected route")
				}
			}))

			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			bosh.SetBodyReader(func(io.Reader) ([]byte, error) {
				return nil, errors.New("a bad read happened")
			})
			_, err := client.DeploymentVMs("some-deployment-name")
			Expect(err).To(MatchError("a bad read happened"))
		})

		It("should error on a bogus response body when unexpected response occurs", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.URL.Path).To(Equal("/deployments/some-deployment-name/vms"))
				w.WriteHeader(http.StatusNotFound)
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
			_, err := client.DeploymentVMs("some-deployment-name")
			Expect(err).To(MatchError("a bad read happened"))
		})
	})
})
