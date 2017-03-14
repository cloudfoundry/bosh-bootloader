package gcpbackend

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
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

type GCPBackend struct {
	handleListInstances func(http.ResponseWriter, *http.Request)
	handlerMutex        sync.Mutex
	Network             Network
}

func (g *GCPBackend) StartFakeGCPBackend() (*httptest.Server, string) {
	fakeGCPServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		g.handlerMutex.Lock()
		defer g.handlerMutex.Unlock()
		switch req.URL.Path {
		case "/o/oauth2/token":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"token": "my-oauth-token"}`))
			return
		case "/some-project-id":
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
		case "/some-project-id/zones/some-zone/instances":
			if g.handleListInstances != nil {
				g.handleListInstances(w, req)
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{}`))
			}
			return
		case "/some-project-id/global/networks":
			w.WriteHeader(http.StatusOK)
			networkName := strings.Split(req.URL.Query().Get("filter"), " ")[2]
			if g.Network.Exists(networkName) {
				w.Write([]byte(`{"items":[{"id":"1234"}]}`))
			} else {
				w.Write([]byte(`{}`))
			}
		default:
			log.Println("unexpected request recieved: ", req.URL.Path)
			w.WriteHeader(http.StatusTeapot)
		}
	}))

	serviceAccountKey := fmt.Sprintf(`{"type": "service_account",
  "project_id": "some-project-id",
  "private_key_id": "private-key-id",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDZGpFTgZpVIO/K\nHTtDumVVlpvAEobq7V9jMElZaX/luLsxnFR5RRr0LIdLhQkDfJ9k0I5E9GBkA8Nk\nCHCoI1uMMRw82W7nC18iE5NpARu1628HvbqXWYHodrj0Awem5krGypoxiJIOCykP\notabiwOk8s8EZRSt/zx5cIRvaPXPstRcqXi580gku8S5NnIKo4UBkaLE3pv1qDI4\naS5YFf3WpVT6b5vlCKZHo9hTcA6YPqO9r0jo9pzhDZEbMNvqs9jM2EgIN4KYqpZb\ndbkLtF1iKyZ/P49HZQ4FXqUCMV6tFNFBmzCOjWRuO0JZ6QjtBsXmgSB5WJ0ICIkd\noCG+ilmrAgMBAAECggEBAMz2eB0OTlXwMnHuBvV6FBEpjwFWfGlukI9kFtuC7mxC\navf7TwTuaPP81f5GKqxQC2tyOd5/mEDUDLN0BGe4ecVw1+fanwkhgz74nEKV+UNW\ncgws4uvgZPTCoPo9ogu/fvkObWQ2Oy1m++z3HwTZyScA1NChXVSnksBTqbREs0zR\nGpELTKSl9iDh8cjr2ADu+a1pasPDY/44xc2lTJ4vMS9w/s8Rls6Ttc5BAPEw9yBl\nYDb9ctMSiu3cFjfukbkxI9IZADu4Qx2jtAGtTAQXbt6p1ooZnwHZKIMJYAsPj0xQ\nIbzHaMNbyqrAJGDGhijoCoXZKFj4yu4sAhkuyMJIXoECgYEA9SPTEi6Pvy5jakZY\nm6nRnEB46OOTk/y0H5E4fmOvvgQd/2khiYsGwZST8hCwkFtHvfhmamzOydWfzgRO\nHU8FjmGqdtKBvNZwfvfPJx90i0OCdIiwetv7ycHCOjGPc3XN1lXvCoNXHhyQ/RpY\n8t8NCjL4kLrZnkdgvQcJWKUndEUCgYEA4rjHtqXsubloGSEeLl93kR5tbgZwiP3u\nj78s3xNVLAisFUk10VdhzP3Ga2HazL9DSos8YfVZBShDcVB2AJQfPgaQ+hpEgc/W\nDfdoBPmKnanHIE3gqwvIa81NOlZcTGqV5mk34FsII//tmciN33DjNercxuGDBGrD\niM5tEh1ajS8CgYAOaIijbPEt/4AAYxoaLCUR1ghFR/sIm7XKlTKI2zsdJAjPVlKO\nTwmany0C8VAva+4PkGYUo0iUPGYkKcSdnGNrNvpZ+Y1+l+wMymv2lLa46MLmLpKQ\n5hUqiqTr3rXbx3TNwEdIiue38V3kQoQv4kRV8SEDALiBwRhChANcnnhvMQKBgC9E\nNKa4etzRcYljpSYn0waXIFtCzm1Q+05One0325bdi/q4E5c8L3CMK7SxZusuqLm+\nw2zsuI1hsoXKL3+5YbYNqmXp2gRyLv8kaDQ5ThPGlHQAqGkggL0wxPv3izCHPA8Y\nOoT0lYLj1UYtUJ6Xq1bPSw3PcAAYvgEkgAq5weoTAoGALuyJJopr2qn2pmaKxA0P\njsySqrsmRr7yteKapm5PfAcYh4BgbiOPg9wKW9rOwPiRaf/y+k7bxpecWM+YNkVu\nSHxWBmTn59QxOSJuOq/U77ZNOuoNptAZe5x/IEIu2kAXiDylq88swnBRp8l6ARtb\nJ/2o9W+sKrGoUyJcA/0Jo7I=\n-----END PRIVATE KEY-----\n",
  "client_email": "some-user@group.iam.gserviceaccount.com",
  "client_id": "123456789",
  "auth_uri": "%[1]s/o/oauth2/auth",
  "token_uri": "%[1]s/o/oauth2/token"
}`, fakeGCPServer.URL)

	return fakeGCPServer, serviceAccountKey
}

func (g *GCPBackend) HandleListInstances(f func(http.ResponseWriter, *http.Request)) {
	g.handlerMutex.Lock()
	defer g.handlerMutex.Unlock()
	g.handleListInstances = f
}
