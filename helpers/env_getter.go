package helpers

import "os"

type envGetter struct {
}

// EnvGetter defines fetching environment variables
type EnvGetter interface {
	Get(name string) string
}

// NewEnvGetter creates a new env getter
func NewEnvGetter() EnvGetter {
	return &envGetter{}
}

func (*envGetter) Get(name string) string {
	return os.Getenv(name)
}
