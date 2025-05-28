package port

import "crypto/tls"

type Certificate interface {
	Get(name string) (*tls.Certificate, error)
	Names() []string
	IsEmpty() bool
}
