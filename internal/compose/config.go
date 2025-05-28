package compose

import (
	"fmt"
	"reflect"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	HTTPPort     uint16 `env:"HTTP_PORT" envDefault:"80"`
	HTTPSPort    uint16 `env:"HTTPS_PORT" envDefault:"443"`
	CertDir      string `env:"CERT_DIR" envDefault:"/etc/letsencrypt/live/"`
	CertFileName string `env:"CERT_FILE_NAME" envDefault:"fullchain.pem"`
	KeyFileName  string `env:"KEY_FILE_NAME" envDefault:"privkey.pem"`
	AllowedHosts string `env:"ALLOWED_HOSTS" envDefault:".*"`
}

func GetConfig() Config {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}

func (c Config) Print() {
	v := reflect.ValueOf(c)

	fmt.Println("Configuration:")
	fmt.Println("----------------------")

	for i := range v.NumField() {
		field := v.Type().Field(i)
		value := v.Field(i)

		fmt.Printf("%-15s: %v\n", field.Name, value.Interface())
	}

	fmt.Println("----------------------")
}
