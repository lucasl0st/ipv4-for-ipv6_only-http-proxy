package adapter

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/lucasl0st/ipv4-for-ipv6_only-http-proxy/internal/port"
)

type certificate struct {
	certs map[string]*tls.Certificate
}

func NewCertificate(certsPath, certFileName, keyFileName string) (port.Certificate, error) {
	c := &certificate{
		certs: map[string]*tls.Certificate{},
	}
	err := c.init(certsPath, certFileName, keyFileName)
	return c, err
}

func (c *certificate) init(certsPath, certFileName, keyFileName string) error {
	certDirs, err := os.ReadDir(certsPath)
	if err != nil {
		return fmt.Errorf("failed to read certs dir: %s", err)
	}

	for _, certDir := range certDirs {
		if !certDir.IsDir() {
			continue
		}

		cert, err := tls.LoadX509KeyPair(
			path.Join(certsPath, certDir.Name(), certFileName),
			path.Join(certsPath, certDir.Name(), keyFileName),
		)
		if err != nil {
			return fmt.Errorf("failed to load cert %s: %s", certDir.Name(), err)
		}

		c.certs[certDir.Name()] = &cert
	}

	return nil
}

func (c *certificate) Get(name string) (*tls.Certificate, error) {
	var foundCert *tls.Certificate = nil
	var foundCertWeight uint

	for _, cert := range c.certs {
		for _, subCert := range cert.Certificate {
			x509Cert, err := x509.ParseCertificate(subCert)
			if err != nil {
				return nil, fmt.Errorf("failed to parse cert: %s", err)
			}

			for _, dnsName := range x509Cert.DNSNames {
				matches := isSubdomain(dnsName, name)
				if !matches {
					continue
				}

				if foundCert == nil {
					foundCert = cert
					foundCertWeight = uint(len(dnsName))
					continue
				}

				if uint(len(dnsName)) > foundCertWeight {
					foundCert = cert
					foundCertWeight = uint(len(dnsName))
				}
			}
		}
	}

	if foundCert == nil {
		return nil, fmt.Errorf("no cert found for %s", name)
	}

	return foundCert, nil
}

func (c *certificate) Names() []string {
	var names []string

	for name := range c.certs {
		names = append(names, name)
	}

	return names
}

func (c *certificate) IsEmpty() bool {
	return len(c.certs) == 0
}

func isSubdomain(dnsName, subdomain string) bool {
	if dnsName == subdomain {
		return true
	}

	if strings.HasPrefix(dnsName, "*.") {
		if subdomain == dnsName[2:] {
			return false
		}

		return strings.HasSuffix(subdomain, dnsName[2:])
	} else if dnsName == subdomain {
		return true
	}

	return false
}
