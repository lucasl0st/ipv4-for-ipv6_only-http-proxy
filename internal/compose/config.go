package compose

import (
	"flag"
	"fmt"
	"reflect"

	"github.com/caarlos0/env/v11"
)

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringSlice) Set(val string) error {
	*s = append(*s, val)
	return nil
}

type Config struct {
	// env
	HTTPPort     uint16 `env:"HTTP_PORT" envDefault:"80"`
	HTTPSPort    uint16 `env:"HTTPS_PORT" envDefault:"443"`
	CertDir      string `env:"CERT_DIR" envDefault:"/etc/letsencrypt/live/"`
	CertFileName string `env:"CERT_FILE_NAME" envDefault:"fullchain.pem"`
	KeyFileName  string `env:"KEY_FILE_NAME" envDefault:"privkey.pem"`
	AllowedHosts string `env:"ALLOWED_HOSTS" envDefault:".*"`

	// flags
	SourceIPFilterHosts stringSlice
	SourceIPFilterCIDRs stringSlice
}

func GetConfig() Config {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		panic(err)
	}

	flag.Var(&cfg.SourceIPFilterHosts, "source-ip-filter-hosts", "")
	flag.Var(&cfg.SourceIPFilterCIDRs, "source-ip-filter-cidrs", "")
	flag.Parse()

	if len(cfg.SourceIPFilterHosts) != len(cfg.SourceIPFilterCIDRs) {
		panic("must provide the same number of hosts to match cidrs")
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
