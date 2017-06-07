package backend

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
)

const (
	defaultVersion = "2.0.0"
)

type backendData struct {
	version             string
	path                string
	createEnvFastFail   bool
	deleteEnvFastFail   bool
	callRealInterpolate bool
	interpolateArgs     []string
	createEnvArgs       string
	createEnvCallCount  int
}

type Backend struct {
	server       *httptest.Server
	handlerMutex sync.Mutex
	backendMutex sync.Mutex
	handler      func(w http.ResponseWriter, r *http.Request)

	backend *backendData
}

func NewBackend() *Backend {
	backend := &Backend{
		backend: &backendData{},
	}
	backend.server = httptest.NewServer(http.HandlerFunc(backend.ServeHTTP))
	backend.ResetAll()

	return backend
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

func (b *Backend) defaultHandler(responseWriter http.ResponseWriter, request *http.Request) {
	switch request.URL.Path {
	case "/version":
		b.handleVersion(responseWriter)
	case "/path":
		b.handlePath(responseWriter)
	case "/interpolate/args":
		b.handleInterpolateArgs(request)
	case "/create-env/args":
		b.handleCreateEnvArgs(request)
	case "/create-env/fastfail":
		b.handleCreateEnvFastFail(responseWriter)
	case "/create-env/call-count":
		b.handleCreateEnvCallCount(responseWriter)
	case "/delete-env/fastfail":
		b.handleDeleteEnvFastFail(responseWriter)
	case "/call-real-interpolate":
		b.handleCallRealInterpolate(responseWriter)
	default:
		responseWriter.WriteHeader(http.StatusOK)
		return
	}
}

func (b *Backend) ResetAll() {
	b.handlerMutex.Lock()
	defer b.handlerMutex.Unlock()

	path := b.backend.path
	b.backend = &backendData{
		version: defaultVersion,
		path:    path,
	}
}

func (b *Backend) SetVersion(version string) {
	b.handlerMutex.Lock()
	defer b.handlerMutex.Unlock()

	b.backend.version = version
}

func (b *Backend) ResetVersion() {
	b.handlerMutex.Lock()
	defer b.handlerMutex.Unlock()

	b.backend.version = defaultVersion
}

func (b *Backend) SetHandler(f func(w http.ResponseWriter, r *http.Request)) {
	b.handlerMutex.Lock()
	defer b.handlerMutex.Unlock()

	b.handler = f
}

func (b *Backend) ServerURL() string {
	return b.server.URL
}

func (b *Backend) SetPath(path string) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	b.backend.path = path
}

func (b *Backend) SetCreateEnvFastFail(fastFail bool) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	b.backend.createEnvFastFail = fastFail
}

func (b *Backend) SetDeleteEnvFastFail(fastFail bool) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	b.backend.deleteEnvFastFail = fastFail
}

func (b *Backend) SetCallRealInterpolate(callRealInterpolate bool) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	b.backend.callRealInterpolate = callRealInterpolate
}

func (b *Backend) ResetInterpolateArgs() {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	b.backend.interpolateArgs = []string{}
}

func (b *Backend) GetInterpolateArgs(index int) string {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	return b.backend.interpolateArgs[index]
}

func (b *Backend) CreateEnvCallCount() int {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	return b.backend.createEnvCallCount
}

func (b *Backend) handleInterpolateArgs(request *http.Request) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	b.backend.interpolateArgs = append(b.backend.interpolateArgs, string(body))
}

func (b *Backend) handleCreateEnvArgs(request *http.Request) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	b.backend.createEnvArgs = string(body)
}

func (b *Backend) handleCreateEnvFastFail(responseWriter http.ResponseWriter) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	if b.backend.createEnvFastFail {
		responseWriter.WriteHeader(http.StatusInternalServerError)
	} else {
		responseWriter.WriteHeader(http.StatusOK)
	}
}

func (b *Backend) handleCreateEnvCallCount(responseWriter http.ResponseWriter) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	b.backend.createEnvCallCount++
}

func (b *Backend) handleDeleteEnvFastFail(responseWriter http.ResponseWriter) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	if b.backend.deleteEnvFastFail {
		responseWriter.WriteHeader(http.StatusInternalServerError)
	} else {
		responseWriter.WriteHeader(http.StatusOK)
	}
}

func (b *Backend) handleCallRealInterpolate(responseWriter http.ResponseWriter) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	if b.backend.callRealInterpolate {
		responseWriter.Write([]byte("true"))
	} else {
		responseWriter.Write([]byte("false"))
	}
}

func (b *Backend) handleVersion(responseWriter http.ResponseWriter) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	responseWriter.Write([]byte(b.backend.version))
}

func (b *Backend) handlePath(responseWriter http.ResponseWriter) {
	b.backendMutex.Lock()
	defer b.backendMutex.Unlock()

	responseWriter.Write([]byte(b.backend.path))
}
