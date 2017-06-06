package cloudconfig

import (
	"io/ioutil"
	"os"

	"golang.org/x/net/proxy"
)

func SetTempDir(f func(string, string) (string, error)) {
	tempDir = f
}

func ResetTempDir() {
	tempDir = ioutil.TempDir
}

func SetWriteFile(f func(string, []byte, os.FileMode) error) {
	writeFile = f
}

func ResetWriteFile() {
	writeFile = ioutil.WriteFile
}

func SetProxySOCKS5(f func(string, string, *proxy.Auth, proxy.Dialer) (proxy.Dialer, error)) {
	proxySOCKS5 = f
}

func ResetProxySOCKS5() {
	proxySOCKS5 = proxy.SOCKS5
}
