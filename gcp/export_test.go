package gcp

import (
	"net/http"

	"golang.org/x/oauth2/jwt"
)

func SetGCPHTTPClient(f func(*jwt.Config) *http.Client) {
	gcpHTTPClient = f
}

func ResetGCPHTTPClient() {
	gcpHTTPClient = gcpHTTPClientFunc
}
