package bosh

import (
	"os"

	"golang.org/x/net/proxy"
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

func SetProxySOCKS5(f func(string, string, *proxy.Auth, proxy.Dialer) (proxy.Dialer, error)) {
	proxySOCKS5 = f
}

func ResetProxySOCKS5() {
	proxySOCKS5 = proxy.SOCKS5
}
