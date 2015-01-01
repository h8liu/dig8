package dns8

import (
	"bytes"
	"fmt"
	"net"
)

// RdIPv6 is a IPv6 rdata
type RdIPv6 net.IP

// UnpackRdIPv6 unpacks an IPv6 record
func UnpackRdIPv6(in *bytes.Reader, n uint16) (RdIPv6, error) {
	if n != 16 {
		return nil, fmt.Errorf("IPv6 with %d bytes", n)
	}
	buf := make([]byte, 16)
	_, e := in.Read(buf)
	if e != nil {
		return nil, e
	}

	return RdIPv6(buf), nil
}

// PrintTo prints the record to output
func (d RdIPv6) PrintTo(out *bytes.Buffer) {
	fmt.Fprint(out, net.IP(d))
}

// Pack packs the record to bytes
func (d RdIPv6) Pack() []byte {
	return net.IP(d).To16()
}
