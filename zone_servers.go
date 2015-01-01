package dig8

import (
	"net"
)

// ZoneServers keep records name servers and their IPs if any
type ZoneServers struct {
	zone       *Domain
	ips        map[uint32]*NameServer
	resolved   map[string]*Domain
	unresolved map[string]*Domain

	records []*RR // related records
}

// Zone returns the zone of the server set
func (zs *ZoneServers) Zone() *Domain { return zs.zone }

// NewZoneServers returns an empty server set for zone.
func NewZoneServers(zone *Domain) *ZoneServers {
	return &ZoneServers{
		zone,
		make(map[uint32]*NameServer),
		make(map[string]*Domain),
		make(map[string]*Domain),
		nil,
	}
}

func (zs *ZoneServers) addUnresolved(server *Domain) bool {
	s := server.String()
	if _, found := zs.unresolved[s]; found {
		return false
	}
	if _, found := zs.resolved[s]; found {
		return false
	}

	zs.unresolved[s] = server
	return true
}

func (zs *ZoneServers) add(server *Domain, ip net.IP) bool {
	index := ipUint(ip)
	if _, found := zs.ips[index]; found {
		return false
	}

	s := server.String()
	if _, found := zs.unresolved[s]; found {
		delete(zs.unresolved, s)
	}

	zs.ips[index] = &NameServer{
		Zone:   zs.zone,
		Domain: server,
		IP:     ip,
	}

	zs.resolved[server.String()] = server

	return true
}

// Add adds a name server with ips into the server set.
// Returns true if any of the ips are new.
func (zs *ZoneServers) Add(server *Domain, ips ...net.IP) bool {
	if len(ips) == 0 {
		return zs.addUnresolved(server)
	}

	anyAdded := false
	for _, ip := range ips {
		if zs.add(server, ip) {
			anyAdded = true
		}
	}

	return anyAdded
}

// ListResolved returns the list of name servers that
// has a resolved IP address.
func (zs *ZoneServers) ListResolved() []*NameServer {
	resolved := make([]*NameServer, 0, len(zs.ips))
	for _, s := range zs.ips {
		resolved = append(resolved, s)
	}

	return resolved
}

// ListUnresolved returns the list of name servers
// that has not a resolved IP address.
func (zs *ZoneServers) ListUnresolved() []*NameServer {
	unresolved := make([]*NameServer, 0, len(zs.unresolved))
	for _, d := range zs.unresolved {
		unresolved = append(unresolved, &NameServer{
			Zone:   zs.zone,
			Domain: d,
			IP:     nil,
		})
	}
	return unresolved
}

// Prepare returns a shuffled list of resolved
// and unresolved name servers.
func (zs *ZoneServers) Prepare() (res, unres []*NameServer) {
	res = shuffleList(zs.ListResolved())
	unres = shuffleList(zs.ListUnresolved())
	return
}

// List returns all the name servers, resolved first.
func (zs *ZoneServers) List() []*NameServer {
	ret := make([]*NameServer, 0, len(zs.ips)+len(zs.unresolved))
	ret = append(ret, zs.ListResolved()...)
	ret = append(ret, zs.ListUnresolved()...)
	return ret
}

// Serves checks if the servers is serving domain d.
func (zs *ZoneServers) Serves(d *Domain) bool {
	return zs.zone.IsZoneOf(d)
}

// Servers get a zone server set from a reply packet from zone for domain d.
// It might print warning messages to the printer if the anything weird
// is detected.
func Servers(p *Packet, z *Domain, d *Domain, pr *Printer) *ZoneServers {
	redirects := p.SelectRedirects(z, d)
	if len(redirects) == 0 {
		return nil
	}

	next := redirects[0].Domain

	ret := NewZoneServers(next)
	ret.records = redirects

	for _, rr := range redirects {
		if !rr.Domain.Equal(next) {
			pr.Printf("// warning: ignore different subzone: %v", rr.Domain)
			continue
		}

		ns := RdToDomain(rr.Rdata)

		rrs := p.SelectIPs(ns) // glued IPs
		ret.records = append(ret.records, rrs...)

		ips := make([]net.IP, 0, len(rrs))
		for _, rr := range rrs {
			ips = append(ips, RdToIPv4(rr.Rdata))
		}
		ret.Add(ns, ips...)
	}

	return ret
}

// Records returns the saved related records.
func (zs *ZoneServers) Records() []*RR { return zs.records }

// AddRecords adds the records to the zone server set as related
// records.
func (zs *ZoneServers) AddRecords(list []*RR) {
	zs.records = append(zs.records, list...)
}
