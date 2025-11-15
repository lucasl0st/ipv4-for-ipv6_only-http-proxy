package service

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"

	"github.com/lucasl0st/ipv4-for-ipv6_only-http-proxy/internal/port"
)

type proxy struct {
	filters []port.Filter
	dns     port.DNS

	transports sync.Map

	maxIdleConnectionsPerHost int
	attemptHTTP2              bool
}

func NewProxy(
	filters []port.Filter,
	dns port.DNS,
	maxIdleConnectionsPerHost int,
	attemptHTTP2 bool,
) http.Handler {
	return &proxy{
		filters:                   filters,
		dns:                       dns,
		maxIdleConnectionsPerHost: maxIdleConnectionsPerHost,
		attemptHTTP2:              attemptHTTP2,
	}
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := strings.Split(r.Host, ":")[0]

	var wErr error
	defer func() {
		if wErr == nil {
			return
		}

		slog.Error("error writing response", "error", wErr, "host", host, "method", r.Method, "url", r.URL)
	}()

	for _, filter := range p.filters {
		if !filter.Filter(w, r) {
			return
		}
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	r.URL.Scheme = scheme
	r.URL.Host = host

	aaaa, err := p.dns.AAAA(host)
	if err != nil {
		slog.Error("could not find ipv6 address", "error", err, "method", r.Method, "url", r.URL)
		w.WriteHeader(http.StatusBadGateway)
		_, wErr = w.Write([]byte("bad gateway"))
		return
	}

	target := fmt.Sprintf("[%s]", *aaaa)
	slog.Info("proxying request", "method", r.Method, "url", r.URL, "target", target)
	r.URL.Host = target

	reverseProxy := httputil.NewSingleHostReverseProxy(r.URL)

	reverseProxy.Transport = p.transportForHost(host)
	reverseProxy.Director = directorForHost(host, r)

	reverseProxy.ServeHTTP(w, r)
}

func (p *proxy) transportForHost(host string) *http.Transport {
	cachedTransport, ok := p.transports.Load(host)
	if ok {
		transport, ok := cachedTransport.(*http.Transport)
		if ok {
			return transport
		}
	}

	transport := &http.Transport{
		MaxIdleConns:        p.maxIdleConnectionsPerHost,
		MaxIdleConnsPerHost: p.maxIdleConnectionsPerHost,
		ForceAttemptHTTP2:   p.attemptHTTP2,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: host,
		},
	}
	p.transports.Store(host, transport)
	return transport
}

func directorForHost(host string, r *http.Request) func(req *http.Request) {
	return func(req *http.Request) {
		req.Host = host
		req.URL = r.URL
	}
}
