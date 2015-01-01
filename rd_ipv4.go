package dig8

import (
	"bytes"
	"fmt"
	"net"
)

// RdIPv4 records an A record
type RdIPv4 net.IP

// UnpackRdIPv4 unpacks the thing
func UnpackRdIPv4(in *bytes.Reader, n uint16) (RdIPv4, error) {
	if n != 4 {
		return nil, fmt.Errorf("IPv4 with %d bytes", n)
	}

	buf := make([]byte, 4)
	_, e := in.Read(buf)
	if e != nil {
		return nil, e
	}

	return RdIPv4(net.IPv4(buf[0], buf[1], buf[2], buf[3])), nil
}

// PrintTo prints the thing to output
func (d RdIPv4) PrintTo(out *bytes.Buffer) {
	fmt.Fprint(out, net.IP(d))
}

// Pack packs the thing
func (d RdIPv4) Pack() []byte {
	return net.IP(d).To4()
}

// RdToIPv4 converts rdata to IPv4 address
func RdToIPv4(r Rdata) net.IP {
	return (net.IP)(r.(RdIPv4))
}
