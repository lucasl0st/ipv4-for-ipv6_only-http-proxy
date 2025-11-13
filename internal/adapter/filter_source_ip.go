package adapter

import (
	"errors"
	"log/slog"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/lucasl0st/ipv4-for-ipv6_only-http-proxy/internal/port"
)

type filterSourceIP struct {
	hosts      regexp.Regexp
	allowedIPs []net.IPNet
}

func NewFilterSourceIP(hosts string, allowedIPs string) (port.Filter, error) {
	var r *regexp.Regexp

	if len(hosts) == 0 {
		return nil, errors.New("no allowed hosts specified")
	}

	var err error
	r, err = regexp.Compile(hosts)
	if err != nil {
		return nil, errors.Join(errors.New("failed to compile ip source filter hosts regex"), err)
	}

	var nets []net.IPNet
	parts := strings.Split(allowedIPs, ",")
	for _, part := range parts {
		_, ipNet, err := net.ParseCIDR(part)
		if err != nil {
			return nil, err
		}

		nets = append(nets, *ipNet)
	}

	return &filterSourceIP{
		hosts:      *r,
		allowedIPs: nets,
	}, nil
}

func (f filterSourceIP) Filter(w http.ResponseWriter, r *http.Request) bool {
	host := strings.Split(r.Host, ":")[0]

	if !f.hosts.MatchString(host) {
		return true
	}

	remoteIP := net.ParseIP(r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")])
	if remoteIP == nil {
		slog.Error("failed to parse remote ip address", "remote", r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
		return false
	}

	for _, ipNet := range f.allowedIPs {
		if ipNet.Contains(remoteIP) {
			return true
		}
	}

	slog.Warn("remote not allowed", "host", host, "method", r.Method, "url", r.URL, "remote", r.RemoteAddr)
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte("bad request"))
	return false
}
