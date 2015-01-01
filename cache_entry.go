package dig8

import (
	"time"
)

const (
	cacheLifeSpan = time.Hour
)

type cacheEntry struct {
	zone       *Domain
	ips        map[uint32]*NameServer
	resolved   map[string]*Domain
	unresolved map[string]*Domain
	expires    time.Time
}

func (e *cacheEntry) Expired() bool {
	return time.Now().After(e.expires)
}

func (e *cacheEntry) addResolved(d *Domain) {
	s := d.String()
	e.resolved[s] = d
	if e.unresolved[s] != nil {
		delete(e.unresolved, s)
	}
}

func emptyCacheEntry(zone *Domain) *cacheEntry {
	return &cacheEntry{
		zone,
		make(map[uint32]*NameServer),
		make(map[string]*Domain),
		make(map[string]*Domain),
		time.Now().Add(cacheLifeSpan),
	}
}

func newCacheEntry(zs *ZoneServers) *cacheEntry {
	ret := emptyCacheEntry(zs.zone)
	ret.Add(zs)
	return ret
}

func (e *cacheEntry) Add(zs *ZoneServers) {
	if !zs.zone.Equal(e.zone) {
		panic("zone mismatch")
	}

	for key, ns := range zs.ips {
		e.ips[key] = ns
		e.addResolved(ns.Domain)
	}

	for key, d := range zs.unresolved {
		s := d.String()
		if key != s {
			panic("bug")
		}

		e.unresolved[s] = d
	}
}

func (e *cacheEntry) ZoneServers() *ZoneServers {
	ret := NewZoneServers(e.zone)

	for _, ns := range e.ips {
		ret.Add(ns.Domain, ns.IP)
	}

	for _, d := range e.unresolved {
		ret.Add(d)
	}

	return ret
}
