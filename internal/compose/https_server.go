package compose

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/lucasl0st/ipv4-for-ipv6_only-http-proxy/internal/port"
)

func newHTTPSServer(handler http.Handler, certificate port.Certificate, port uint16) *http.Server {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return certificate.Get(info.ServerName)
		},
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second,
		TLSConfig:         tlsConfig,
		Handler:           handler,
	}

	return server
}
