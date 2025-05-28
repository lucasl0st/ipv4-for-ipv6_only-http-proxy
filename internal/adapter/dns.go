package adapter

import (
	"fmt"
	"net"

	"github.com/lucasl0st/ipv4-for-ipv6_only-http-proxy/internal/port"
)

type dns struct{}

func NewDNS() port.DNS {
	return &dns{}
}

func (d *dns) AAAA(host string) (*string, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}

	var aaaa string

	for _, ip := range ips {
		if ip.To4() == nil && ip.To16() != nil {
			aaaa = ip.String()
			break
		}
	}

	if len(aaaa) == 0 {
		return nil, fmt.Errorf("could not find AAAA record for %s", host)
	}

	return &aaaa, nil
}
