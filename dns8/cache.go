package dns8

import (
	"time"
)

// Cache is a name server cache
type Cache struct {
	RegistrarOnly bool
	entries       map[string]*cacheEntry

	puts chan *cachePut
	gets chan *cacheGet
}

type cachePut struct {
	zs    *ZoneServers
	reply chan bool
}

type cacheGet struct {
	zone  *Domain
	reply chan *ZoneServers
}

// NewCache creates a new name server cache
func NewCache() *Cache {
	ret := new(Cache)
	ret.entries = make(map[string]*cacheEntry)
	ret.RegistrarOnly = true

	ret.puts = make(chan *cachePut)
	ret.gets = make(chan *cacheGet)

	go ret.serve()

	return ret
}

func (c *Cache) serve() {
	ticker := time.Tick(time.Minute * 5)

	for {
		select {
		case put := <-c.puts:
			put.reply <- c.put(put.zs)
		case get := <-c.gets:
			get.reply <- c.get(get.zone)
		case <-ticker:
			c.clean()
		}
	}
}

func (c *Cache) put(z *ZoneServers) bool {
	zone := z.Zone()
	if zone.IsRoot() {
		return false // we never cache root
	}
	if c.RegistrarOnly && !zone.IsRegistrar() {
		return false
	}

	key := zone.String()
	entry := c.entries[key]
	if entry == nil {
		c.entries[key] = newCacheEntry(z)
	} else {
		entry.Add(z)
	}

	return true
}

func (c *Cache) cleanZone(z *Domain) {
	zstr := z.String()
	entry := c.entries[z.String()]
	if entry == nil {
		return
	}

	if entry.Expired() {
		delete(c.entries, zstr)
	}
}

func (c *Cache) clean() {
	toClean := make([]string, 0, 100)
	for k, v := range c.entries {
		if v.Expired() {
			toClean = append(toClean, k)
		}
	}

	for _, k := range toClean {
		delete(c.entries, k)
	}
}

func (c *Cache) get(z *Domain) *ZoneServers {
	c.cleanZone(z)

	entry := c.entries[z.String()]
	if entry == nil {
		return nil
	}

	return entry.ZoneServers()
}

// Get queries the zone servers for a domain
func (c *Cache) Get(z *Domain) *ZoneServers {
	ch := make(chan *ZoneServers)
	c.gets <- &cacheGet{z, ch}
	return <-ch
}

// Put puts the zone servers into the cache.
// Returns true if the put is successful.
// Returns false if the put is rejected.
func (c *Cache) Put(zs *ZoneServers) bool {
	ch := make(chan bool)
	c.puts <- &cachePut{zs, ch}
	return <-ch
}
