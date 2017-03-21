package backend

import (
	"net/http"
	"net/http/httptest"
	"sync"
)

type Backend struct {
	server *httptest.Server

	handlerMutex sync.Mutex
	handler      func(w http.ResponseWriter, r *http.Request)
}

func NewBackend() *Backend {
	backend := &Backend{}
	backend.server = httptest.NewServer(http.HandlerFunc(backend.ServeHTTP))
	return backend
}

func (b *Backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.handlerMutex.Lock()
	defer b.handlerMutex.Unlock()
	if b.handler != nil {
		b.handler(w, r)
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
