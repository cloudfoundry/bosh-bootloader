package gcp_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	compute "google.golang.org/api/compute/v1"

	"github.com/cloudfoundry/bosh-bootloader/gcp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/oauth2/jwt"
)

const (
	GetProjectOutput = `{"commonInstanceMetadata": {
    "items": [ { "key": "sshKeys", "value": "user:ssh-rsa something user" } ], "kind": "compute#metadata"
  },
  "name": "some-project-name",
  "quotas": [ { "limit": 25000, "metric": "SNAPSHOTS" } ],
  "selfLink": "https://www.googleapis.com/compute/v1/projects/some-project-id"
}`
	SetCommonInstanceMetadataOutput = `{
 "kind": "compute#operation",
 "name": "operation-12345",
 "operationType": "setMetadata",
 "targetId": "12345",
 "status": "PENDING",
 "user": "user@example.com",
 "selfLink": "https://www.googleapis.com/compute/v1/projects/cf-release-integration/global/operations/operation-1478888342819-5410a865610b9-fa8ffd77-0d4332fc"
 }`
)

var _ = Describe("ProjectsService", func() {
	var (
		fakeGCPServer     *httptest.Server
		serviceAccountKey string
		getProjectsCall   int
		provider          *gcp.Provider
	)

	BeforeEach(func() {
		fakeGCPServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			switch req.URL.Path {
			case "/o/oauth2/token":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"token": "my-oauth-token"}`))
				return
			case "/some-project-id":
				getProjectsCall++
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(GetProjectOutput))
				return
			case "/some-project-id/setCommonInstanceMetadata":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(SetCommonInstanceMetadataOutput))
				return
			case "/invalid-project-id/setCommonInstanceMetadata":
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"code":1,"errorSpace":"core","status":403,"message":"invalid-project"}`))
				return
			case "/non-existing-project-id":
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"code":1,"errorSpace":"core","status":403,"message":"forbidden"}`))
				return
			default:
				log.Println("unexpected request recieved: ", req.URL.Path)
				w.WriteHeader(http.StatusTeapot)
			}
		}))

		serviceAccountKey = fmt.Sprintf(`{"type": "service_account",
"project_id": "some-project-id",
"private_key_id": "private-key-id",
"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDZGpFTgZpVIO/K\nHTtDumVVlpvAEobq7V9jMElZaX/luLsxnFR5RRr0LIdLhQkDfJ9k0I5E9GBkA8Nk\nCHCoI1uMMRw82W7nC18iE5NpARu1628HvbqXWYHodrj0Awem5krGypoxiJIOCykP\notabiwOk8s8EZRSt/zx5cIRvaPXPstRcqXi580gku8S5NnIKo4UBkaLE3pv1qDI4\naS5YFf3WpVT6b5vlCKZHo9hTcA6YPqO9r0jo9pzhDZEbMNvqs9jM2EgIN4KYqpZb\ndbkLtF1iKyZ/P49HZQ4FXqUCMV6tFNFBmzCOjWRuO0JZ6QjtBsXmgSB5WJ0ICIkd\noCG+ilmrAgMBAAECggEBAMz2eB0OTlXwMnHuBvV6FBEpjwFWfGlukI9kFtuC7mxC\navf7TwTuaPP81f5GKqxQC2tyOd5/mEDUDLN0BGe4ecVw1+fanwkhgz74nEKV+UNW\ncgws4uvgZPTCoPo9ogu/fvkObWQ2Oy1m++z3HwTZyScA1NChXVSnksBTqbREs0zR\nGpELTKSl9iDh8cjr2ADu+a1pasPDY/44xc2lTJ4vMS9w/s8Rls6Ttc5BAPEw9yBl\nYDb9ctMSiu3cFjfukbkxI9IZADu4Qx2jtAGtTAQXbt6p1ooZnwHZKIMJYAsPj0xQ\nIbzHaMNbyqrAJGDGhijoCoXZKFj4yu4sAhkuyMJIXoECgYEA9SPTEi6Pvy5jakZY\nm6nRnEB46OOTk/y0H5E4fmOvvgQd/2khiYsGwZST8hCwkFtHvfhmamzOydWfzgRO\nHU8FjmGqdtKBvNZwfvfPJx90i0OCdIiwetv7ycHCOjGPc3XN1lXvCoNXHhyQ/RpY\n8t8NCjL4kLrZnkdgvQcJWKUndEUCgYEA4rjHtqXsubloGSEeLl93kR5tbgZwiP3u\nj78s3xNVLAisFUk10VdhzP3Ga2HazL9DSos8YfVZBShDcVB2AJQfPgaQ+hpEgc/W\nDfdoBPmKnanHIE3gqwvIa81NOlZcTGqV5mk34FsII//tmciN33DjNercxuGDBGrD\niM5tEh1ajS8CgYAOaIijbPEt/4AAYxoaLCUR1ghFR/sIm7XKlTKI2zsdJAjPVlKO\nTwmany0C8VAva+4PkGYUo0iUPGYkKcSdnGNrNvpZ+Y1+l+wMymv2lLa46MLmLpKQ\n5hUqiqTr3rXbx3TNwEdIiue38V3kQoQv4kRV8SEDALiBwRhChANcnnhvMQKBgC9E\nNKa4etzRcYljpSYn0waXIFtCzm1Q+05One0325bdi/q4E5c8L3CMK7SxZusuqLm+\nw2zsuI1hsoXKL3+5YbYNqmXp2gRyLv8kaDQ5ThPGlHQAqGkggL0wxPv3izCHPA8Y\nOoT0lYLj1UYtUJ6Xq1bPSw3PcAAYvgEkgAq5weoTAoGALuyJJopr2qn2pmaKxA0P\njsySqrsmRr7yteKapm5PfAcYh4BgbiOPg9wKW9rOwPiRaf/y+k7bxpecWM+YNkVu\nSHxWBmTn59QxOSJuOq/U77ZNOuoNptAZe5x/IEIu2kAXiDylq88swnBRp8l6ARtb\nJ/2o9W+sKrGoUyJcA/0Jo7I=\n-----END PRIVATE KEY-----\n",
"client_email": "some-user@group.iam.gserviceaccount.com",
"client_id": "123456789",
"auth_uri": "%[1]s/o/oauth2/auth",
"token_uri": "%[1]s/o/oauth2/token"
}`, fakeGCPServer.URL)

		provider = gcp.NewProvider(fakeGCPServer.URL)
	})

	Describe("SetCommonInstanceMetadata", func() {
		var projectsService gcp.ProjectsService

		BeforeEach(func() {
			err := provider.SetConfig(serviceAccountKey)
			Expect(err).NotTo(HaveOccurred())

			serviceWrapper := provider.GetService()
			projectsService = serviceWrapper.GetProjectsService()
		})

		It("updates the metadata for the given project", func() {
			someValue := "some-random-value"
			metadata := &compute.Metadata{
				Items: []*compute.MetadataItems{
					{
						Key:   "some-random-key",
						Value: &someValue,
					},
				},
			}
			operation, err := projectsService.SetCommonInstanceMetadata("some-project-id", metadata)
			Expect(err).NotTo(HaveOccurred())
			Expect(operation.OperationType).To(Equal("setMetadata"))
		})

		Context("failure cases", func() {
			It("returns an error when the update fails", func() {
				metadata := &compute.Metadata{}

				_, err := projectsService.SetCommonInstanceMetadata("invalid-project-id", metadata)
				Expect(err).To(MatchError(`googleapi: got HTTP response code 403 with body: {"code":1,"errorSpace":"core","status":403,"message":"invalid-project"}`))
			})
		})
	})

	Describe("Get", func() {
		It("gets the project details for a given project-id", func() {
			err := provider.SetConfig(serviceAccountKey)
			Expect(err).NotTo(HaveOccurred())

			serviceWrapper := provider.GetService()
			projectService := serviceWrapper.GetProjectsService()

			project, err := projectService.Get("some-project-id")
			Expect(err).NotTo(HaveOccurred())
			Expect(getProjectsCall).To(Equal(1))
			Expect(project.Name).To(Equal("some-project-name"))
			Expect(project.CommonInstanceMetadata.Items).To(HaveLen(1))
		})

		Context("failure cases", func() {
			AfterEach(func() {
				gcp.ResetClient()
			})

			It("returns an error when the service account key is not valid json", func() {
				err := provider.SetConfig("1231:123")
				Expect(err).To(MatchError("invalid character ':' after top-level value"))
			})

			It("returns an error when a service could not be created", func() {
				gcp.SetClient(func(*jwt.Config) *http.Client {
					return nil
				})
				err := provider.SetConfig(`{"type": "service_account"}`)
				Expect(err).To(MatchError("client is nil"))
			})

			It("returns an error when the project could not be found", func() {
				err := provider.SetConfig(serviceAccountKey)
				Expect(err).NotTo(HaveOccurred())

				serviceWrapper := provider.GetService()
				projectService := serviceWrapper.GetProjectsService()

				_, err = projectService.Get("non-existing-project-id")
				Expect(err).To(MatchError(`googleapi: got HTTP response code 403 with body: {"code":1,"errorSpace":"core","status":403,"message":"forbidden"}`))
			})
		})
	})
})
