package main

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

var (
	allowedHosts regexp.Regexp
	dns          *DNS
	certs        *CertStore
)

func main() {
	var cfg config
	err := env.Parse(&cfg)
	if err != nil {
		slog.Error("failed to parse config", "error", err)
		os.Exit(1)
	}

	cfg.Print()

	certs, err = NewCertStore(cfg.CertDir, cfg.CertFileName, cfg.KeyFileName)
	if err != nil {
		slog.Error("failed to initialize cert store", "error", err)
		os.Exit(1)
	}

	if certs.IsEmpty() {
		slog.Error("no certs found in cert dir", "certDir", cfg.CertDir)
		os.Exit(1)
	}

	dns = NewDNS(cfg.CacheDNS, cfg.DNSCacheTTL)

	if len(cfg.AllowedHosts) == 0 {
		panic("no allowed hosts specified")
	} else {
		if cfg.AllowedHosts == ".*" {
			slog.Warn("allowing all hosts, this is insecure!")
		}

		r, err := regexp.Compile(cfg.AllowedHosts)
		if err != nil {
			slog.Error("failed to compile allowed hosts regex", "error", err)
			os.Exit(1)
		}

		allowedHosts = *r
	}

	go listenHTTP(cfg)
	go listenHTTPs(cfg)

	select {}
}

func listenHTTP(cfg config) {
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           http.HandlerFunc(handler),
	}

	slog.Info("starting http server", "addr", server.Addr)

	err := server.ListenAndServe()
	if err != nil {
		slog.Error("http server failed", "error", err)
		os.Exit(1)
	}
}

func listenHTTPs(cfg config) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return certs.Get(info.ServerName)
		},
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPSPort),
		ReadHeaderTimeout: 3 * time.Second,
		TLSConfig:         tlsConfig,
		Handler:           http.HandlerFunc(handler),
	}
	slog.Info("starting https server", "addr", server.Addr)

	initialCert := certs.Names()[0]
	cert := path.Join(cfg.CertDir, initialCert, cfg.CertFileName)
	key := path.Join(cfg.CertDir, initialCert, cfg.KeyFileName)

	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		slog.Error("https server failed", "error", err)
		os.Exit(1)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	host := strings.Split(r.Host, ":")[0]

	if !allowedHosts.MatchString(host) {
		slog.Warn("host not allowed", "host", host, "method", r.Method, "url", r.URL)

		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("bad request"))
		if err != nil {
			slog.Error("failed to write response", "error", err, "method", r.Method, "url", r.URL)
		}
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	r.URL.Scheme = scheme
	r.URL.Host = host

	aaaa, err := dns.AAAA(host)
	if err != nil {
		slog.Error("could not find ipv6 address", "error", err, "method", r.Method, "url", r.URL)
		w.WriteHeader(http.StatusBadGateway)
		_, err = w.Write([]byte("bad gateway"))
		if err != nil {
			slog.Error("failed to write response", "error", err, "method", r.Method, "url", r.URL)
		}
		return
	}

	target := fmt.Sprintf("[%s]", *aaaa)

	slog.Info("proxying request", "method", r.Method, "url", r.URL, "target", target)

	r.URL.Host = target

	proxy := httputil.NewSingleHostReverseProxy(r.URL)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: host,
		},
	}

	proxy.Transport = tr

	proxy.Director = func(req *http.Request) {
		req.Host = host
		req.URL = r.URL
	}

	proxy.ServeHTTP(w, r)
}
