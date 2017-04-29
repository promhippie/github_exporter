package exporter

import (
	"crypto/tls"
	"net/http"

	"github.com/jackspirou/syscerts"
)

// simpleClient initializes a simple HTTP client.
func simpleClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				RootCAs: syscerts.SystemRootsPool(),
			},
		},
	}
}
