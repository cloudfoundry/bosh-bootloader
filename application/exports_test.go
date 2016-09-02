package application

import (
	"os"

	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

func SetGetwd(f func() (string, error)) {
	getwd = f
}

func ResetGetwd() {
	getwd = os.Getwd
}

func SetGetState(f func(string) (storage.State, error)) {
	getState = f
}

func ResetGetState() {
	getState = storage.GetState
}
