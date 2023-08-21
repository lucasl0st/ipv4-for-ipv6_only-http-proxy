package main

type config struct {
	HttpPort  uint16 `env:"HTTP_PORT" envDefault:"80"`
	HttpsPort uint16 `env:"HTTPS_PORT" envDefault:"443"`
	CertFile  string `env:"CERT_FILE" envDefault:"cert.pem"`
	KeyFile   string `env:"KEY_FILE" envDefault:"key.pem"`
}
