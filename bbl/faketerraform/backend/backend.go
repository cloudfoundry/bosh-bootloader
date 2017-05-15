package backend

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
)

const (
	defaultVersion = "0.8.6"
)

type dataBackend struct {
	fakeBOSHServerURL     string
	version               string
	outputJsonReturnError bool
	fastFail              bool
}

type Backend struct {
	server *httptest.Server

	backendMutex sync.Mutex
	handlerMutex sync.Mutex
	handler      func(w http.ResponseWriter, r *http.Request)

	backend *dataBackend
}

func NewBackend() *Backend {
	backend := &Backend{
		backend: &dataBackend{},
	}
	backend.server = httptest.NewServer(http.HandlerFunc(backend.ServeHTTP))
	backend.ResetAll()

	return backend
}

func (b *Backend) ResetAll() {
	b.backend = &dataBackend{
		version: defaultVersion,
	}
}

func (b *Backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.handlerMutex.Lock()
	defer b.handlerMutex.Unlock()

	if b.handler != nil {
		b.handler(w, r)
	} else {
		b.defaultHandler(w, r)
	}
}

func (b *Backend) SetFakeBOSHServer(url string) {
	b.backend.fakeBOSHServerURL = url
}

func (b *Backend) defaultHandler(responseWriter http.ResponseWriter, request *http.Request) {
	switch request.URL.Path {
	case "/output/--json":
		b.handleOutputJson(responseWriter)
	case "/output/external_ip":
		responseWriter.Write([]byte("127.0.0.1"))
	case "/output/director_address":
		responseWriter.Write([]byte(b.backend.fakeBOSHServerURL))
	case "/output/network_name":
		b.handleOutput(responseWriter, "some-network-name")
	case "/output/subnetwork_name":
		b.handleOutput(responseWriter, "some-subnetwork-name")
	case "/output/internal_tag_name":
		b.handleOutput(responseWriter, "some-internal-tag")
	case "/output/bosh_open_tag_name":
		b.handleOutput(responseWriter, "some-bosh-tag")
	case "/output/concourse_target_pool":
		b.handleOutput(responseWriter, "concourse-target-pool")
	case "/output/router_backend_service":
		b.handleOutput(responseWriter, "router-backend-service")
	case "/output/ssh_proxy_target_pool":
		b.handleOutput(responseWriter, "ssh-proxy-target-pool")
	case "/output/tcp_router_target_pool":
		b.handleOutput(responseWriter, "tcp-router-target-pool")
	case "/output/ws_target_pool":
		b.handleOutput(responseWriter, "ws-target-pool")
	case "/output/router_lb_ip":
		b.handleOutput(responseWriter, "some-router-lb-ip")
	case "/output/ssh_proxy_lb_ip":
		b.handleOutput(responseWriter, "some-ssh-proxy-lb-ip")
	case "/output/tcp_router_lb_ip":
		b.handleOutput(responseWriter, "some-tcp-router-lb-ip")
	case "/output/concourse_lb_ip":
		b.handleOutput(responseWriter, "some-concourse-lb-ip")
	case "/output/ws_lb_ip":
		b.handleOutput(responseWriter, "some-ws-lb-ip")
	case "/output/system_domain_dns_servers":
		b.handleOutput(responseWriter, "name-server-1.,\nname-server-2.,\nname-server-3.")
	case "/fastfail":
		b.handleFastFail(responseWriter)
	case "/version":
		b.handleVersion(responseWriter)
	}
}

func (b *Backend) handleOutput(responseWriter http.ResponseWriter, output string) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	if b.backend.outputJsonReturnError {
		responseWriter.WriteHeader(http.StatusInternalServerError)
	} else {
		responseWriter.Write([]byte(output))
	}
}

func (b *Backend) handleOutputJson(responseWriter http.ResponseWriter) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	if b.backend.outputJsonReturnError {
		responseWriter.WriteHeader(http.StatusInternalServerError)
	} else {
		responseWriter.Write([]byte(fmt.Sprintf(`{
			"bosh_eip": {
				"value": "some-bosh-eip"
			},
			"bosh_url": {
				"value": %q
			},
			"bosh_user_access_key": {
				"value": "some-bosh-user-access-key"
			},
			"bosh_user_secret_access_key": {
				"value": "some-bosh-user-secret-access_key"
			},
			"nat_eip": {
				"value": "some-nat-eip"
			},
			"bosh_subnet_id": {
				"value": "some-bosh-subnet-id"
			},
			"bosh_subnet_availability_zone": {
				"value": "some-bosh-subnet-availability-zone"
			},
			"bosh_security_group": {
				"value": "some-bosh-security-group"
			},
			"env_dns_zone_name_servers": {
				"value": [
					"name-server-1.",
					"name-server-2."
				]
			},
			"internal_security_group": {
				"value": "some-internal-security-group"
			},
			"internal_subnet_ids": {
				"value": [
					"some-internal-subnet-ids-1",
					"some-internal-subnet-ids-2",
					"some-internal-subnet-ids-3"
				]
			},
			"internal_subnet_cidrs": {
				"value": [
					"10.0.16.0/20",
					"10.0.32.0/20",
					"10.0.48.0/20"
				]
			},
			"vpc_id": {
				"value": "some-vpc-id"
			},
			"cf_router_lb_name": {
				"value": "some-cf-router-lb"
			},
			"cf_router_lb_url": {
				"value": "some-cf-router-lb-url"
			},
			"cf_router_lb_internal_security_group": {
				"value": "some-cf-router-internal-security-group"
			},
			"cf_ssh_lb_name":  {
				"value": "some-cf-ssh-proxy-lb"
			},
			"cf_ssh_lb_url": {
				"value": "some-cf-ssh-proxy-lb-url"
			},
			"cf_ssh_lb_internal_security_group":  {
				"value": "some-cf-ssh-proxy-internal-security-group"
			},
			"concourse_lb_name":  {
				"value": "some-concourse-lb"
			},
			"concourse_lb_internal_security_group":  {
				"value": "some-concourse-internal-security-group"
			}
		}`, b.backend.fakeBOSHServerURL)))
	}
}

func (b *Backend) SetOutputJsonReturnError(errorOut bool) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	b.backend.outputJsonReturnError = errorOut
}

func (b *Backend) SetVersion(version string) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	b.backend.version = version
}

func (b *Backend) ResetVersion() {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	b.backend.version = defaultVersion
}

func (b *Backend) handleVersion(responseWriter http.ResponseWriter) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	responseWriter.Write([]byte(b.backend.version))
}

func (b *Backend) SetFastFail(fastFail bool) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	b.backend.fastFail = fastFail
}

func (b *Backend) handleFastFail(responseWriter http.ResponseWriter) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	if b.backend.fastFail {
		responseWriter.WriteHeader(http.StatusInternalServerError)
	} else {
		responseWriter.WriteHeader(http.StatusOK)
	}
}

func (b *Backend) SetHandler(f func(w http.ResponseWriter, r *http.Request)) {
	b.handlerMutex.Lock()
	defer b.handlerMutex.Unlock()

	b.handler = f
}

func (b *Backend) ServerURL() string {
	return b.server.URL
}
