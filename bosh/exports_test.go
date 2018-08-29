package bosh

import (
	"os"
)

func SetOSSetenv(f func(string, string) error) {
	osSetenv = f
}

func ResetOSSetenv() {
	osSetenv = os.Setenv
}

func SetOSUnsetenv(f func(string) error) {
	osUnsetenv = f
}

func ResetOSUnsetenv() {
	osUnsetenv = os.Unsetenv
}
