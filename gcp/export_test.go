package gcp

import (
	"net/http"

	"golang.org/x/oauth2/jwt"
)

func SetClient(f func(*jwt.Config) *http.Client) {
	client = f
}

func ResetClient() {
	client = clientFunc
}
