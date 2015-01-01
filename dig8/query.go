package dig8

import (
	"fmt"
	"net"
	"time"
)

// Query is a query to a name server.
type Query struct {
	Domain *Domain
	Type   uint16
	Server *net.UDPAddr

	Zone       *Domain
	ServerName *Domain
}

// Server converts an IP address to UDP address with DNSPort
func Server(ip net.IP) *net.UDPAddr {
	return &net.UDPAddr{
		IP:   ip,
		Port: DNSPort,
	}
}

// Q makes a new query
func Q(d *Domain, t uint16, at net.IP) *Query {
	return &Query{
		Domain: d,
		Type:   t,
		Server: Server(at),
	}
}

// Qs makes a query from string domains and IPs.
func Qs(d string, t uint16, at string) *Query {
	return Q(D(d), t, net.ParseIP(at))
}

func (q *Query) addrString() string {
	if q.ServerName == nil {
		return addrString(q.Server)
	}

	p := q.Server.Port
	if p == 0 || p == DNSPort {
		return fmt.Sprintf("%v(%v)", q.ServerName, q.Server.IP)
	}
	return fmt.Sprintf("%v(%v):%d", q.ServerName, q.Server.IP, p)
}

func (q *Query) String() string {
	return fmt.Sprintf("%v %s @%s",
		q.Domain,
		TypeString(q.Type),
		q.addrString(),
	)
}

func newMessage(q *Query, id uint16) *Message {
	return &Message{
		RemoteAddr: q.Server,
		Packet:     QpackID(q.Domain, q.Type, id),
		Timestamp:  time.Now(),
	}
}
