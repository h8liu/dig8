package dig8

import (
	"net"
)

type ipsResult struct {
	cnames  []*RR
	results []*RR
}

// IPs is a query task for querying the IP address for
// a particular domain
type IPs struct {
	Domain     *Domain
	StartWith  *ZoneServers
	HeadLess   bool
	HideResult bool

	// inherit from the initializing Recur Task
	Return  int
	Packet  *Packet
	EndWith *ZoneServers
	Zones   []*ZoneServers

	CnameTraceBack map[string]*Domain // in and out, inherit from father IPs

	CnameEndpoints []*Domain       // new endpoint cnames discovered
	CnameIPs       map[string]*IPs // sub IPs for each unresolved end point

	CnameRecords []*RR // new cname records
	Records      []*RR // new end point ip records

	resultSave *ipsResult
}

// NewIPs creates a new query task for IPs.
func NewIPs(d *Domain) *IPs {
	return &IPs{Domain: d}
}

// collectResults look for Query error or A records in Answer
func (ips *IPs) collectResults(recur *Recur) {
	if recur.Return != Okay {
		panic("bug")
	}

	for _, rr := range recur.Answers {
		switch rr.Type {
		case A:
			ips.Records = append(ips.Records, rr)
		case CNAME:
			// okay
		default:
			panic("bug")
		}
	}
}

func (ips *IPs) findCnameResults(recur *Recur) (unresolved []*Domain) {
	unresolved = make([]*Domain, 0, len(ips.CnameEndpoints))

	for _, cname := range ips.CnameEndpoints {
		rrs := recur.Packet.SelectRecords(cname, A)
		if len(rrs) == 0 {
			unresolved = append(unresolved, cname)
			continue
		}
		ips.Records = append(ips.Records, rrs...)
	}

	return
}

// Returns true when if finds any endpoints
func (ips *IPs) extractCnames(recur *Recur, d *Domain, c Cursor) bool {
	if _, found := ips.CnameTraceBack[d.String()]; !found {
		panic("bug")
	}

	if !ips.EndWith.Serves(d) {
		// domain not in the zone
		// so even there were cname records about this domain
		// they cannot be trusted
		return false
	}

	rrs := recur.Packet.SelectRecords(d, CNAME)
	ret := false

	for _, rr := range rrs {
		cname := RdToDomain(rr.Rdata)
		cnameStr := cname.String()
		if ips.CnameTraceBack[cnameStr] != nil {
			// some error cnames, pointing to self or forming circles
			continue
		}

		c.P().Printf("// cname: %v -> %v", d, cname)
		ips.CnameRecords = append(ips.CnameRecords, rr)
		ips.CnameTraceBack[cname.String()] = d

		// see if it follows another CNAME
		if ips.extractCnames(recur, cname, c) {
			// see so, then we only tracks the end point
			ret = true // we added an endpoint in the recursion
			continue
		}

		c.P().Printf("// cname endpoint: %v", cname)
		// these are end points that needs to be crawled
		ips.CnameEndpoints = append(ips.CnameEndpoints, cname)
		ret = true
	}

	return ret
}

// PrintResult prints the results using the cursor.
func (ips *IPs) PrintResult(c Cursor) {
	cnames, results := ips.Results()
	p := c.P()

	for _, r := range cnames {
		p.Printf("// %v -> %v", r.Domain, RdToDomain(r.Rdata))
	}

	if len(results) == 0 {
		p.Printf("// (%v is unresolvable)", ips.Domain)
	}

	for _, r := range results {
		p.Printf("// %v(%v)", r.Domain, RdToIPv4(r.Rdata))
	}
}

// Run queries for the IP addresses.
func (ips *IPs) Run(c Cursor) {
	p := c.P()

	if !ips.HeadLess {
		p.Printf("ips %v {", ips.Domain)
		p.ShiftIn()
		defer p.ShiftOut("}")
	}

	ips.run(c)
	if c.E() != nil {
		return
	}

	if !ips.HideResult {
		ips.PrintResult(c)
	}
}

// Results returns the cnames and the ip records.
func (ips *IPs) Results() (cnames, results []*RR) {
	if ips.resultSave != nil {
		return ips.resultSave.cnames, ips.resultSave.results
	}

	cnames = make([]*RR, 0, 20)
	results = make([]*RR, 0, 20)
	cnames, results = ips.results(cnames, results)
	ips.resultSave = &ipsResult{cnames, results}

	return
}

// ResultAndIPs returns the records and the ip addresses
func (ips *IPs) ResultAndIPs() (cnames, res []*RR, retIPs []net.IP) {
	cnames, res = ips.Results()
	if len(res) == 0 {
		return
	}

	hits := make(map[uint32]bool)
	retIPs = make([]net.IP, 0, len(res))

	for _, rr := range res {
		ip := RdToIPv4(rr.Rdata)
		index := ipUint(ip)
		if hits[index] {
			continue
		}
		hits[index] = true
		retIPs = append(retIPs, ip)
	}

	return
}

// IPs returns the IP addresses.
func (ips *IPs) IPs() []net.IP {
	_, _, ret := ips.ResultAndIPs()
	return ret
}

func (ips *IPs) results(cnames, results []*RR) (c, r []*RR) {
	cnames = append(cnames, ips.CnameRecords...)
	results = append(results, ips.Records...)

	for _, cnameIPs := range ips.CnameIPs {
		cnames, results = cnameIPs.results(cnames, results)
	}

	return cnames, results
}

func (ips *IPs) run(c Cursor) {
	recur := NewRecur(ips.Domain)
	recur.HeadLess = true
	recur.StartWith = ips.StartWith

	_, e := c.T(recur)
	if e != nil {
		return
	}

	// inherit from recur
	ips.Return = recur.Return
	ips.EndWith = recur.EndWith
	ips.Packet = recur.Packet
	ips.Zones = recur.Zones

	if ips.Return != Okay {
		return
	}

	ips.Records = make([]*RR, 0, 10)
	ips.collectResults(recur)

	// even if we find results, we still track cnames if any
	ips.CnameEndpoints = make([]*Domain, 0, 10)
	if ips.CnameTraceBack == nil {
		ips.CnameTraceBack = make(map[string]*Domain)
		ips.CnameTraceBack[ips.Domain.String()] = nil
	} else {
		_, found := ips.CnameTraceBack[ips.Domain.String()]
		if !found {
			panic("bug")
		}
	}

	ips.CnameRecords = make([]*RR, 0, 10)
	if !ips.extractCnames(recur, ips.Domain, c) {
		return
	}

	if len(ips.CnameEndpoints) == 0 {
		panic("bug")
	}

	unresolved := ips.findCnameResults(recur)
	if len(unresolved) == 0 {
		return
	}

	// trace down the cnames
	p := ips.Packet
	z := ips.EndWith
	ips.CnameIPs = make(map[string]*IPs)

	for _, cname := range unresolved {
		// search for redirects
		servers := Servers(p, z.Zone(), cname, c.P())

		// check for last result
		if servers == nil {
			if z.Serves(cname) {
				servers = z
			}
		}

		if servers == nil {
			if ips.StartWith != nil && ips.StartWith.Serves(cname) {
				servers = ips.StartWith
			}
		}

		cnameIPs := NewIPs(cname)
		cnameIPs.HideResult = true
		cnameIPs.StartWith = servers
		cnameIPs.CnameTraceBack = ips.CnameTraceBack

		ips.CnameIPs[cname.String()] = cnameIPs

		_, e := c.T(cnameIPs)
		if e != nil {
			return
		}
	}
}

// PrintTo prints the task via the printer
func (ips *IPs) PrintTo(p *Printer) {
	cnames, results := ips.Results()

	if len(cnames) > 0 {
		for _, r := range cnames {
			p.Printf("%v -> %v", r.Domain, RdToDomain(r.Rdata))
		}
		p.Println()
	}

	if len(results) == 0 {
		p.Print("(unresolvable)")
	} else {

		for _, r := range results {
			d := r.Domain
			ip := RdToIPv4(r.Rdata)
			if d.Equal(ips.Domain) {
				p.Printf("%v", ip)
			} else {
				p.Printf("%v(%v)", ip, d)
			}
		}
	}
}

func init() {
	nsResolve = func(c Cursor, d *Domain, zs *ZoneServers) ([]net.IP, error) {
		t := NewIPs(d)
		if _, e := c.T(t); e != nil {
			return nil, e
		}

		cnames, res, ips := t.ResultAndIPs()
		zs.AddRecords(cnames)
		zs.AddRecords(res)
		zs.Add(d, ips...)

		return ips, nil
	}
}
