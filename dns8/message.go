package dns8

import (
	"net"
	"time"
)

// Message is a DNS message with a remote address
type Message struct {
	RemoteAddr *net.UDPAddr
	Packet     *Packet
	Timestamp  time.Time
}

func addrString(a *net.UDPAddr) string {
	if a.Port == 0 || a.Port == DNSPort {
		return a.IP.String()
	}
	return a.String()
}

// PrintTo prints it the message to a printer
func (m *Message) PrintTo(p *Printer) {
	p.Printf("@%s", addrString(m.RemoteAddr))
	m.Packet.PrintTo(p)
}

func (m *Message) String() string {
	return PrintStr(m)
}
