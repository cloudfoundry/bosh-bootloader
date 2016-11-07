package commands

import "os"

type EnvGetter struct{}

func NewEnvGetter() EnvGetter {
	return EnvGetter{}
}

func (EnvGetter) Get(name string) string {
	return os.Getenv(name)
}
