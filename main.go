package main

import (
	"crypto/tls"
	"fmt"
	"github.com/caarlos0/env/v9"
	"log"
	"net/http"
	"net/http/httputil"
	"path"
	"regexp"
	"strings"
)

var allowedHosts regexp.Regexp
var dns *Dns
var certs *CertStore

func main() {
	var cfg config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("failed to parse config: %s", err)
	}

	cfg.Print()

	certs, err = NewCertStore(cfg.CertDir, cfg.CertFileName, cfg.KeyFileName)
	if err != nil {
		log.Fatalf("failed to initialize cert store: %s", err)
	}

	if certs.IsEmpty() {
		log.Fatalf("no certs found in %s", cfg.CertDir)
	}

	dns = NewDns(cfg.CacheDNS, cfg.DNSCacheTTL)

	if len(cfg.AllowedHosts) == 0 {
		panic("no allowed hosts specified")
	} else {
		if cfg.AllowedHosts == ".*" {
			fmt.Println("allowing all hosts, this is insecure!")
		}

		r, err := regexp.Compile(cfg.AllowedHosts)
		if err != nil {
			log.Fatalf("failed to compile allowed hosts regex: %s", err)
		}

		allowedHosts = *r
	}

	go listenHttp(cfg)
	go listenHttps(cfg)

	select {}
}

func listenHttp(cfg config) {
	log.Printf("http server started on port 80\n")

	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.HttpPort), http.HandlerFunc(handler))
	if err != nil {
		log.Fatal(err)
	}
}

func listenHttps(cfg config) {
	tlsConfig := &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			return certs.Get(info.ServerName)
		},
	}

	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", cfg.HttpsPort),
		TLSConfig: tlsConfig,
		Handler:   http.HandlerFunc(handler),
	}

	log.Printf("https server started on port 443\n")

	initialCert := certs.Names()[0]
	cert := path.Join(cfg.CertDir, initialCert, cfg.CertFileName)
	key := path.Join(cfg.CertDir, initialCert, cfg.KeyFileName)

	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	host := strings.Split(r.Host, ":")[0]

	if !allowedHosts.MatchString(host) {
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
	r.URL.Host = host

	aaaa, err := dns.AAAA(host)
	if err != nil {
		log.Printf("%s: %s %v", "could not find ipv6 address", r.Method, r.URL)
		w.WriteHeader(http.StatusBadGateway)
		_, err = w.Write([]byte("bad gateway"))
		if err != nil {
			log.Printf("%s: %s %v", err, r.Method, r.URL)
		}
		return
	}

	target := fmt.Sprintf("[%s]", *aaaa)

	log.Printf("%s %v", r.Method, r.URL)

	r.URL.Host = target

	proxy := httputil.NewSingleHostReverseProxy(r.URL)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{ServerName: host},
	}

	proxy.Transport = tr

	proxy.Director = func(req *http.Request) {
		req.Host = host
		req.URL = r.URL
	}

	proxy.ServeHTTP(w, r)
}
