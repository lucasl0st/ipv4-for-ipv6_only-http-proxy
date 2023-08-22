package main

import (
	"fmt"
	"reflect"
)

type config struct {
	HttpPort     uint16 `env:"HTTP_PORT" envDefault:"80"`
	HttpsPort    uint16 `env:"HTTPS_PORT" envDefault:"443"`
	CertFile     string `env:"CERT_FILE" envDefault:"cert.pem"`
	KeyFile      string `env:"KEY_FILE" envDefault:"key.pem"`
	AllowedHosts string `env:"ALLOWED_HOSTS" envDefault:".*"`
	CacheDNS     bool   `env:"CACHE_DNS" envDefault:"true"`
	DNSCacheTTL  uint16 `env:"DNS_CACHE_TTL" envDefault:"60"`
}

func (c config) Print() {
	v := reflect.ValueOf(c)

	fmt.Println("Configuration:")
	fmt.Println("----------------------")

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		value := v.Field(i)

		fmt.Printf("%-15s: %v\n", field.Name, value.Interface())
	}

	fmt.Println("----------------------")
}
