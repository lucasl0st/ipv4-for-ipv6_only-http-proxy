package adapter

import (
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/lucasl0st/ipv4-for-ipv6_only-http-proxy/internal/port"
)

type filterAllowedHost struct {
	allowedHosts regexp.Regexp
}

func NewFilterAllowedHost(allowedHosts string) (port.Filter, error) {
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

	return &filterAllowedHost{allowedHosts: *r}, nil
}

func (f filterAllowedHost) Filter(w http.ResponseWriter, r *http.Request) bool {
	host := strings.Split(r.Host, ":")[0]

	if !f.allowedHosts.MatchString(host) {
		slog.Warn("host not allowed", "host", host, "method", r.Method, "url", r.URL)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
		return false
	}

	return true
}
