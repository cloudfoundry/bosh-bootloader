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

var _ = Describe("GetTaskOutput", func() {
	var server *httptest.Server

	BeforeEach(func() {
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.URL.Path).To(Equal("/tasks/1/output"))
			Expect(r.URL.RawQuery).To(Equal("type=event"))
			Expect(r.Method).To(Equal("GET"))

			w.Write([]byte(`
				{"time": 0, "error": {"code": 100, "message": "some-error" }, "stage": "some-stage", "tags": [ "some-tag" ], "total": 1, "task": "some-task-guid", "index": 1, "state": "some-state", "progress": 0}
{"time": 1, "error": {"code": 100, "message": "some-error" }, "stage": "some-stage", "tags": [ "some-tag" ], "total": 1, "task": "some-task-guid", "index": 1, "state": "some-new-state", "progress": 0}
				`))
		}))
	})

	It("returns task event output for a given task", func() {
		client := bosh.NewClient(bosh.Config{
			URL:      server.URL,
			Username: "some-username",
			Password: "some-password",
		})

		taskOutputs, err := client.GetTaskOutput(1)
		Expect(err).NotTo(HaveOccurred())
		Expect(taskOutputs).To(ConsistOf(
			bosh.TaskOutput{
				Time: 0,
				Error: bosh.TaskError{
					Code:    100,
					Message: "some-error",
				},
				Stage:    "some-stage",
				Tags:     []string{"some-tag"},
				Total:    1,
				Task:     "some-task-guid",
				Index:    1,
				State:    "some-state",
				Progress: 0,
			},
			bosh.TaskOutput{
				Time: 1,
				Error: bosh.TaskError{
					Code:    100,
					Message: "some-error",
				},
				Stage:    "some-stage",
				Tags:     []string{"some-tag"},
				Total:    1,
				Task:     "some-task-guid",
				Index:    1,
				State:    "some-new-state",
				Progress: 0,
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

			_, err := client.GetTaskOutput(1)
			Expect(err).To(MatchError(ContainSubstring("invalid URL escape")))
		})

		It("error on an empty URL", func() {
			client := bosh.NewClient(bosh.Config{
				URL:      "",
				Username: "some-username",
				Password: "some-password",
			})

			_, err := client.GetTaskOutput(1)
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

			_, err := client.GetTaskOutput(1)
			Expect(err).To(MatchError("unexpected response 502 Bad Gateway:\nMore Info"))
		})

		It("should error on a bogus response body", func() {
			client := bosh.NewClient(bosh.Config{
				URL:      server.URL,
				Username: "some-username",
				Password: "some-password",
			})

			bosh.SetBodyReader(func(io.Reader) ([]byte, error) {
				return nil, errors.New("a bad read happened")
			})

			_, err := client.GetTaskOutput(1)
			Expect(err).To(MatchError("a bad read happened"))
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

			_, err := client.GetTaskOutput(1)
			Expect(err).To(MatchError(ContainSubstring("invalid character")))
		})
	})
})
