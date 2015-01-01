package dig8

// TODO: packet unpacking might requires a special helper.
// the current one is a little bit messy.

import (
	"bytes"
)

func unpackRdata(t, c uint16, in *bytes.Reader, p []byte) (Rdata, error) {
	n := uint16(in.Len())
	if c == IN {
		switch t {
		case A:
			return UnpackRdIPv4(in, n)
		case NS, CNAME:
			return UnpackRdDomain(in, n, p)
		case AAAA:
			return UnpackRdIPv6(in, n)
		case TXT:
			return UnpackRdTxt(in, n)
		case MX:
			return UnpackRdMx(in, n, p)
		case SOA:
			return UnpackRdSoa(in, n, p)
		}
	}
	return UnpackRdBytes(in, n)
}

// UnpackRdata unpacks an Rdata record from packet buffer
// of type t and code c. p is the original packet for seaching tabs.
func UnpackRdata(t, c uint16, in *bytes.Reader, p []byte) (Rdata, error) {
	buf := make([]byte, 2)
	if _, e := in.Read(buf); e != nil {
		return nil, e
	}
	n := enc.Uint16(buf) // number of bytes

	buf = make([]byte, n)
	if _, e := in.Read(buf); e != nil {
		return nil, e
	}

	in = bytes.NewReader(buf)
	ret, e := unpackRdata(t, c, bytes.NewReader(buf), p)
	return ret, e
}
