package main

import (
	"fmt"
	"reflect"
)

type config struct {
	HTTPPort     uint16 `env:"HTTP_PORT" envDefault:"80"`
	HTTPSPort    uint16 `env:"HTTPS_PORT" envDefault:"443"`
	CertDir      string `env:"CERT_DIR" envDefault:"/etc/letsencrypt/live/"`
	CertFileName string `env:"CERT_FILE_NAME" envDefault:"fullchain.pem"`
	KeyFileName  string `env:"KEY_FILE_NAME" envDefault:"privkey.pem"`
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
