package dig8

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Info is a query task that gets all the related records.
type Info struct {
	Domain     *Domain
	StartWith  *ZoneServers
	HeadLess   bool
	Shallow    bool
	HideResult bool

	EndWith *ZoneServers

	Cnames  []*RR
	Results []*RR

	Records    []*RR
	RecordsMap map[string]*RR

	NameServers    []*NameServer
	NameServersMap map[string]*NameServer

	Zones map[string]*ZoneServers
}

// NewInfo creates a query task that queries all the
// related records for a domain.
func NewInfo(d *Domain) *Info {
	return &Info{Domain: d}
}

// Run executes the info task, queries for all the
// related records using the cursor.
func (info *Info) Run(c Cursor) {
	p := c.P()
	if !info.HeadLess {
		p.Printf("info %v {", info.Domain)
		p.ShiftIn()
		defer p.ShiftOut("}")
	}

	ips := info.run(c)
	if c.E() != nil {
		return
	}

	if !info.HideResult {
		ips.PrintResult(c)

		if len(info.NameServers) > 0 {
			p.Print()
			for _, ns := range info.NameServers {
				p.Printf("// %v", ns)
			}
		}

		if len(info.Records) > 0 {
			p.Print()
			for _, rr := range info.Records {
				p.Printf("// %s", rr.Digest())
			}
		}
	}
}

func (info *Info) appendAll(rrs []*RR) {
	for _, rr := range rrs {
		k := rr.Digest()
		if info.RecordsMap[k] != nil {
			continue
		}
		info.RecordsMap[k] = rr
		info.Records = append(info.Records, rr)
	}
}

func (info *Info) run(c Cursor) *IPs {
	ips := NewIPs(info.Domain)
	ips.StartWith = info.StartWith
	ips.HideResult = true

	_, e := c.T(ips)
	if e != nil {
		return nil
	}

	info.EndWith = ips.EndWith

	info.Cnames, info.Results = ips.Results()

	info.RecordsMap = make(map[string]*RR)
	info.Records = make([]*RR, 0, 100)
	info.Zones = make(map[string]*ZoneServers)
	info.NameServers = make([]*NameServer, 0, 100)
	info.NameServersMap = make(map[string]*NameServer)

	info.appendAll(info.Cnames)
	info.appendAll(info.Results)

	info.collectInfo(ips)

	for _, z := range info.Zones {
		info.queryZone(z, c)
	}

	return ips
}

var infoTypes = []uint16{NS, MX, SOA, TXT}

func (info *Info) collectInfo(ips *IPs) {
	info._collectInfo(ips)

	if info.Shallow {
		return
	}

	for _, ips := range ips.CnameIPs {
		info._collectInfo(ips)
	}
}

func (info *Info) _collectInfo(ips *IPs) {
	for _, z := range ips.Zones {
		if z.Zone().IsRegistrar() {
			continue
		}

		for _, s := range z.List() {
			if s.IP == nil {
				continue
			}
			k := s.Key()
			if info.NameServersMap[k] != nil {
				continue
			}
			info.NameServersMap[k] = s
			info.NameServers = append(info.NameServers, s)
		}

		info.appendAll(z.Records())

		zoneStr := z.Zone().String()
		if info.Zones[zoneStr] == nil {
			info.Zones[zoneStr] = z
		}
	}
}

func (info *Info) queryZone(z *ZoneServers, c Cursor) error {
	for _, t := range infoTypes {
		recur := NewRecurType(z.Zone(), t)
		recur.StartWith = z
		_, e := c.T(recur)
		if e != nil {
			return e
		}

		info.appendAll(recur.Answers)
	}
	return nil
}

// PrintTo prints the info out via the printer.
func (info *Info) PrintTo(p *Printer) {
	if len(info.Cnames) > 0 {
		p.Print("cnames {")
		p.ShiftIn()
		for _, r := range info.Cnames {
			p.Printf("%v -> %v", r.Domain, RdToDomain(r.Rdata))
		}
		p.ShiftOut("}")
	}

	if len(info.Results) == 0 {
		p.Print("(unresolvable)")
	} else {
		p.Print("ips {")
		p.ShiftIn()

		for _, r := range info.Results {
			d := r.Domain
			ip := RdToIPv4(r.Rdata)
			if d.Equal(info.Domain) {
				p.Printf("%v", ip)
			} else {
				p.Printf("%v(%v)", ip, d)
			}
		}

		p.ShiftOut("}")
	}

	if len(info.NameServers) > 0 {
		p.Print("servers {")
		p.ShiftIn()

		for _, ns := range info.NameServers {
			p.Printf("%v", ns)
		}

		p.ShiftOut("}")
	}

	if len(info.Records) > 0 {
		p.Print("records {")
		p.ShiftIn()

		for _, rr := range info.Records {
			p.Printf("%v", rr.Digest())
		}

		p.ShiftOut("}")
	}
}

// Out gets the output of the info task
func (info *Info) Out() string {
	ret := new(bytes.Buffer)
	p := NewPrinter(ret)
	info.PrintTo(p)
	return ret.String()
}

type infoResult struct {
	domain  string
	ips     []string
	cnames  []string
	servers []string
	records []string
}

func newInfoResult(domain string) *infoResult {
	ret := new(infoResult)
	ret.domain = domain
	ret.ips = make([]string, 0, 10)
	ret.cnames = make([]string, 0, 10)
	ret.servers = make([]string, 0, 10)
	ret.records = make([]string, 0, 10)
	return ret
}

func jmarsh(v interface{}) []byte {
	ret, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	return ret
}

// Result gets the result string of the info task.
// It is a tab separated json object list.
func (info *Info) Result() string {
	ret := newInfoResult(info.Domain.String())

	for _, r := range info.Cnames {
		s := fmt.Sprintf("%v -> %v", r.Domain, RdToDomain(r.Rdata))
		ret.cnames = append(ret.cnames, s)
	}

	for _, r := range info.Results {
		d := r.Domain
		ip := RdToIPv4(r.Rdata)
		var s string
		if d.Equal(info.Domain) {
			s = fmt.Sprintf("%v", ip)
		} else {
			s = fmt.Sprintf("%v(%v)", ip, d)
		}

		ret.ips = append(ret.ips, s)
	}

	for _, ns := range info.NameServers {
		s := fmt.Sprintf("%v", ns)
		ret.servers = append(ret.servers, s)
	}
	for _, rr := range info.Records {
		s := fmt.Sprintf("%v", rr.Digest())
		ret.records = append(ret.records, s)
	}

	out := new(bytes.Buffer)

	wr := func(obj interface{}, post string) {
		out.Write(jmarsh(obj))
		out.WriteString(post)
	}

	wr(ret.domain, "\t")
	wr(ret.ips, "\t")
	wr(ret.cnames, "\t")
	wr(ret.servers, "\t")
	wr(ret.records, "\n")

	return out.String()
}
