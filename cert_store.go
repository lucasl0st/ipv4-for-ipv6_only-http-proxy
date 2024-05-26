package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
)

type CertStore struct {
	sync.RWMutex

	path         string
	certFileName string
	keyFileName  string

	certs map[string]*tls.Certificate
}

func NewCertStore(path, certFileName, keyFileName string) (*CertStore, error) {
	certStore := &CertStore{
		path:         path,
		certFileName: certFileName,
		keyFileName:  keyFileName,
	}

	err := certStore.initialize()
	if err != nil {
		return nil, err
	}

	return certStore, nil
}

func (c *CertStore) initialize() error {
	c.Lock()
	defer c.Unlock()

	c.certs = map[string]*tls.Certificate{}

	certDirs, err := os.ReadDir(c.path)
	if err != nil {
		return fmt.Errorf("failed to read certs dir: %s", err)
	}

	for _, certDir := range certDirs {
		if !certDir.IsDir() {
			continue
		}

		cert, err := tls.LoadX509KeyPair(
			path.Join(c.path, certDir.Name(), c.certFileName),
			path.Join(c.path, certDir.Name(), c.keyFileName),
		)
		if err != nil {
			return fmt.Errorf("failed to load cert %s: %s", certDir.Name(), err)
		}

		c.certs[certDir.Name()] = &cert
	}

	return nil
}

func (c *CertStore) Get(name string) (*tls.Certificate, error) {
	c.RLock()
	defer c.RUnlock()

	var foundCert *tls.Certificate = nil
	var foundCertWeight uint

	for _, cert := range c.certs {
		for _, certB := range cert.Certificate {
			x509Cert, err := x509.ParseCertificate(certB)
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

func (c *CertStore) Names() []string {
	c.RLock()
	defer c.RUnlock()

	var names []string

	for name := range c.certs {
		names = append(names, name)
	}

	return names
}

func (c *CertStore) IsEmpty() bool {
	c.RLock()
	defer c.RUnlock()

	return len(c.certs) == 0
}

func isSubdomain(dnsName, subdomain string) bool {
	if dnsName == subdomain {
		return true
	}

	if strings.HasPrefix(dnsName, "*.") {
		return strings.HasSuffix(subdomain, dnsName[2:])
	}

	return false
}
