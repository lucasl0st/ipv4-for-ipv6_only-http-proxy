package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type DNS struct {
	lock sync.Mutex

	shouldCache bool

	cache map[string]CacheEntry
	ttl   uint16
}

type CacheEntry struct {
	AAAA      string
	Timestamp time.Time
}

func NewDNS(cache bool, ttl uint16) *DNS {
	d := &DNS{
		shouldCache: cache,
		cache:       map[string]CacheEntry{},
		ttl:         ttl,
	}

	go d.cacheWorker()
	return d
}

func (d *DNS) cacheWorker() {
	for {
		d.lock.Lock()

		cacheCopy := map[string]CacheEntry{}
		for host, entry := range d.cache {
			cacheCopy[host] = entry
		}

		d.lock.Unlock()

		for host, entry := range cacheCopy {
			_, err := d.lookupAAAA(host)
			if err != nil && entry.Timestamp.Add(time.Second*time.Duration(d.ttl)).Before(time.Now()) {
				d.lock.Lock()
				delete(d.cache, host)
				d.lock.Unlock()
			}
		}

		time.Sleep(time.Second * 10)
	}
}

func (d *DNS) AAAA(host string) (*string, error) {
	d.lock.Lock()

	cached, ok := d.cache[host]
	if ok {
		d.lock.Unlock()

		aaaa := cached.AAAA
		return &aaaa, nil
	}

	d.lock.Unlock()

	cacheEntry, err := d.lookupAAAA(host)
	if err != nil {
		return nil, err
	}

	aaaa := cacheEntry.AAAA
	return &aaaa, nil
}

func (d *DNS) lookupAAAA(host string) (*CacheEntry, error) {
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

	entry := CacheEntry{
		AAAA:      aaaa,
		Timestamp: time.Now(),
	}

	if d.shouldCache {
		d.lock.Lock()
		defer d.lock.Unlock()
		d.cache[host] = entry
	}

	return &entry, nil
}
