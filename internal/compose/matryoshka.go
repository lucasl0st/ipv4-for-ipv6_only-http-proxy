package compose

import (
	"log/slog"
	"os"
	"path"

	"github.com/lucasl0st/ipv4-for-ipv6_only-http-proxy/internal/adapter"
	"github.com/lucasl0st/ipv4-for-ipv6_only-http-proxy/internal/service"
)

func ListenAndServe() {
	cfg := GetConfig()
	cfg.Print()

	dns := adapter.NewDNS()
	certificate, err := adapter.NewCertificate(cfg.CertDir, cfg.CertFileName, cfg.KeyFileName)
	if err != nil {
		panic(err)
	}

	if certificate.IsEmpty() {
		slog.Error("no certificates found", "certDir", cfg.CertDir)
		os.Exit(1)
	}

	proxy, err := service.NewProxy(cfg.AllowedHosts, dns)
	if err != nil {
		panic(err)
	}

	httpServer := newHTTPServer(proxy, cfg.HTTPPort)
	httpSServer := newHTTPSServer(proxy, certificate, cfg.HTTPSPort)

	initialCert := certificate.Names()[0]
	cert := path.Join(cfg.CertDir, initialCert, cfg.CertFileName)
	key := path.Join(cfg.CertDir, initialCert, cfg.KeyFileName)

	go func() {
		err := httpServer.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		err := httpSServer.ListenAndServeTLS(cert, key)
		if err != nil {
			panic(err)
		}
	}()

	select {}
}
