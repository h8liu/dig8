package dig8

import (
	"net"
)

var nsResolve func(c Cursor, d *Domain, zs *ZoneServers) ([]net.IP, error)

var recurCache = NewCache()

// Recur is a recursive query task that searches
// for the domain and type that starts with a zone servers
type Recur struct {
	Domain    *Domain
	Type      uint16
	StartWith *ZoneServers
	HeadLess  bool

	Return  int          // valid when Error is not null
	Packet  *Packet      // valid when Return is Okay
	EndWith *ZoneServers // valid when Return is Okay
	Answers []*RR        // the records in Packet that ends the query
	Zones   []*ZoneServers

	zone *ZoneServers
}

// NewRecur creates a new recursive query for the
// domain's IP address.
func NewRecur(d *Domain) *Recur {
	return NewRecurType(d, A)
}

// NewRecurType creates a new recursive query for the
// domain's record of type t.
func NewRecurType(d *Domain, t uint16) *Recur {
	return &Recur{
		Domain: d,
		Type:   t,
	}
}

var _ Task = new(Recur)

var roots = MakeRoots()

// Reply code
const (
	Working = iota
	Okay
	NotExists // domain not exists
	Lost      // no valid server reachable
)

func (r *Recur) begin() *ZoneServers {
	if r.StartWith != nil {
		return r.StartWith
	}

	cached := recurCache.Get(r.Domain.Registrar())
	if cached != nil {
		return cached
	}

	return roots
}

// Run executes the recursive query using the cursor.
func (r *Recur) Run(c Cursor) {
	p := c.P()
	if !r.HeadLess {
		p.Printf("recur %v %s {", r.Domain, TypeString(r.Type))
		p.ShiftIn()
		defer p.ShiftOut("}")
	}

	r.zone = r.begin()
	r.Zones = make([]*ZoneServers, 0, 100)

	for r.zone != nil {
		next, e := r.query(c)
		if e != nil {
			return
		}

		recurCache.Put(r.zone)
		r.zone = next
	}
}

func (r *Recur) q(c Cursor, ip net.IP, s *Domain) (*ZoneServers, error) {
	q := &Query{
		Domain:     r.Domain,
		Type:       r.Type,
		Server:     Server(ip),
		Zone:       r.zone.Zone(),
		ServerName: s,
	}

	reply, e := c.Q(q)
	if e != nil {
		return nil, e // some resource limit reached
	}

	attempt := reply.Last()

	if attempt.Error != nil {
		c.P().Printf("// unreachable: %v, last error %v", s, attempt.Error)
		return nil, nil
	}

	p := attempt.Recv.Packet

	rcode := p.Rcode()
	if !(rcode == RcodeOkay || rcode == RcodeNameError) {
		c.P().Printf("// server error %s, rcode=%d", s, rcode)
	}

	ans := p.SelectAnswers(r.Domain, r.Type)
	if len(ans) > 0 {
		r.Return = Okay
		r.Packet = p
		r.Answers = ans
		r.EndWith = r.zone

		return nil, nil
	}

	next := Servers(p, r.zone.Zone(), r.Domain, c.P())
	if next == nil {
		r.Return = NotExists
		c.P().Print("// record does not exist")
	}

	return next, nil
}

func (r *Recur) query(c Cursor) (*ZoneServers, error) {
	zone := r.zone
	r.Zones = append(r.Zones, zone)
	resolved, unresolved := zone.Prepare()

	c.P().Printf("// zone: %v", zone.Zone())

	// try resolved servers first
	for _, server := range resolved {
		next, e := r.q(c, server.IP, server.Domain)
		if e != nil || next != nil || r.Return != Working {
			return next, e
		}
	}

	if nsResolve != nil {
		// when all resolved failed, we try unresolved ones
		for _, server := range unresolved {
			bugOn(server.IP != nil)

			ips, e := nsResolve(c, server.Domain, zone)
			if e != nil {
				return nil, e
			}

			for _, ip := range ips {
				next, e := r.q(c, ip, server.Domain)
				if e != nil || next != nil || r.Return != Working {
					return next, e
				}
			}
		}
	}

	c.P().Print("// no reachable server")
	r.Return = Lost
	r.EndWith = zone
	return nil, nil
}

// PrintTo prints the task via the printer.
func (r *Recur) PrintTo(p *Printer) {
	panic("todo") // TODO:
}
