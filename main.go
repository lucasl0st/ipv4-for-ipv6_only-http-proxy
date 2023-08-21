package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	go func() {
		log.Printf("Server started on port 80\n")

		err := http.ListenAndServe(":80", http.HandlerFunc(handler))
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		log.Printf("Server started on port 443\n")

		err := http.ListenAndServeTLS(":443", "cert.pem", "key.pem", http.HandlerFunc(handler))
		if err != nil {
			log.Fatal(err)
		}
	}()

	select {}
}

func handler(w http.ResponseWriter, r *http.Request) {
	ips, err := net.LookupIP(r.Host)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
	}

	var target string

	for _, ip := range ips {
		if ip.To4() == nil && ip.To16() != nil {
			target = fmt.Sprintf("[%s]", ip.String())
			break
		}
	}

	if len(target) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("No IPv6 address found")
		return
	}

	var targetURL string

	if r.TLS != nil {
		targetURL = fmt.Sprintf("https://%s%s", target, r.URL.Path)
	} else {
		targetURL = fmt.Sprintf("http://%s%s", target, r.URL.Path)
	}

	url, err := url.Parse(targetURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("Error parsing target URL:", err)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	proxy.Transport = tr

	proxy.Director = func(req *http.Request) {
		req.Host = r.Host
		req.URL.Scheme = url.Scheme
		req.URL.Host = target
	}

	proxy.ServeHTTP(w, r)
}
