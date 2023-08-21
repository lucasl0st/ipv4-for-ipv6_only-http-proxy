package main

import (
	"crypto/tls"
	"fmt"
	"github.com/caarlos0/env/v9"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"regexp"
)

var allowedHosts regexp.Regexp

func main() {
	var cfg config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("failed to parse config: %s", err)
	}

	if len(cfg.AllowedHosts) == 0 {
		log.Fatalf("no allowed hosts specified, exiting\n")
	} else {
		if cfg.AllowedHosts == ".*" {
			log.Printf("allowing all hosts, this is insecure!\n")
		}

		r, err := regexp.Compile(cfg.AllowedHosts)
		if err != nil {
			log.Fatalf("failed to compile allowed hosts regex: %s", err)
		}

		allowedHosts = *r
		log.Printf("allowed hosts: %s\n", cfg.AllowedHosts)
	}

	go func() {
		log.Printf("http server started on port 80\n")

		err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.HttpPort), http.HandlerFunc(handler))
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		log.Printf("https server started on port 443\n")

		err := http.ListenAndServeTLS(fmt.Sprintf(":%d", cfg.HttpsPort), cfg.CertFile, cfg.KeyFile, http.HandlerFunc(handler))
		if err != nil {
			log.Fatal(err)
		}
	}()

	select {}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if !allowedHosts.MatchString(r.Host) {
		log.Printf("%s: %s %v", "host not allowed", r.Method, r.URL)
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("bad request"))
		if err != nil {
			log.Printf("%s: %s %v", err, r.Method, r.URL)
		}
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	r.URL.Scheme = scheme
	r.URL.Host = r.Host

	ips, err := net.LookupIP(r.Host)
	if err != nil {
		log.Printf("%s: %s %v", err, r.Method, r.URL)
		w.WriteHeader(http.StatusBadGateway)
		_, err = w.Write([]byte("bad gateway"))
		if err != nil {
			log.Printf("%s: %s %v", err, r.Method, r.URL)
		}
		return
	}

	var target string

	for _, ip := range ips {
		if ip.To4() == nil && ip.To16() != nil {
			target = fmt.Sprintf("[%s]", ip.String())
			break
		}
	}

	if len(target) == 0 {
		log.Printf("%s: %s %v", "could not find ipv6 address", r.Method, r.URL)
		w.WriteHeader(http.StatusBadGateway)
		_, err = w.Write([]byte("bad gateway"))
		if err != nil {
			log.Printf("%s: %s %v", err, r.Method, r.URL)
		}
		return
	}

	log.Printf("%s %v", r.Method, r.URL)

	r.URL.Host = target

	proxy := httputil.NewSingleHostReverseProxy(r.URL)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{ServerName: r.Host},
	}

	proxy.Transport = tr

	proxy.Director = func(req *http.Request) {
		req.Host = r.Host
		req.URL = r.URL
	}

	proxy.ServeHTTP(w, r)
}
