package application

import "os"

func SetGetwd(f func() (string, error)) {
	getwd = f
}

func ResetGetwd() {
	getwd = os.Getwd
}
