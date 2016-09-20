package application

import (
	"os"

	"github.com/cloudfoundry/bosh-bootloader/storage"
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
