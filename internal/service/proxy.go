package service

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"

	"github.com/lucasl0st/ipv4-for-ipv6_only-http-proxy/internal/port"
)

type proxy struct {
	allowedHosts regexp.Regexp
	dns          port.DNS
}

func NewProxy(
	allowedHosts string,
	dns port.DNS,
) (http.Handler, error) {
	var r *regexp.Regexp

	if len(allowedHosts) == 0 {
		return nil, errors.New("no allowed hosts specified")
	}

	if allowedHosts == ".*" {
		slog.Warn("allowing all hosts, this is insecure!")
	}

	var err error
	r, err = regexp.Compile(allowedHosts)
	if err != nil {
		return nil, errors.Join(errors.New("failed to compile allowed hosts regex"), err)
	}

	return &proxy{
		allowedHosts: *r,
		dns:          dns,
	}, nil
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

	if !p.allowedHosts.MatchString(host) {
		slog.Warn("host not allowed", "host", host, "method", r.Method, "url", r.URL)
		w.WriteHeader(http.StatusBadRequest)
		_, wErr = w.Write([]byte("bad request"))
		return
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

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: host,
		},
	}
	reverseProxy.Transport = tr
	reverseProxy.Director = func(req *http.Request) {
		req.Host = host
		req.URL = r.URL
	}

	reverseProxy.ServeHTTP(w, r)
}
