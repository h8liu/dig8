package dns8

import (
	"bytes"
	"fmt"
)

// RR is a record
type RR struct {
	Domain *Domain
	Type   uint16
	Class  uint16
	TTL    uint32
	Rdata  Rdata
}

func (rr *RR) packFlags(out *bytes.Buffer) {
	var buf [8]byte
	enc.PutUint16(buf[0:2], rr.Type)
	enc.PutUint16(buf[2:4], rr.Class)
	enc.PutUint32(buf[4:8], rr.TTL)
	out.Write(buf[:])
}

func (rr *RR) pack(out *bytes.Buffer) {
	rr.Domain.Pack(out)
	rr.packFlags(out)
	PackRdata(out, rr.Rdata)
}

func (rr *RR) unpackFlags(in *bytes.Reader) error {
	var buf [8]byte
	if _, e := in.Read(buf[:]); e != nil {
		return e
	}
	rr.Type = enc.Uint16(buf[0:2])
	rr.Class = enc.Uint16(buf[2:4])
	rr.TTL = enc.Uint32(buf[4:8])

	return nil
}

func (rr *RR) unpackRdata(in *bytes.Reader, p []byte) error {
	var e error
	rr.Rdata, e = UnpackRdata(rr.Type, rr.Class, in, p)
	return e
}

func (rr *RR) unpack(in *bytes.Reader, p []byte) error {
	var e error

	rr.Domain, e = UnpackDomain(in, p)
	if e != nil {
		return e
	}

	if e = rr.unpackFlags(in); e != nil {
		return e
	}

	return rr.unpackRdata(in, p)
}

func unpackRR(in *bytes.Reader, p []byte) (*RR, error) {
	ret := new(RR)
	return ret, ret.unpack(in, p)
}

func (rr *RR) String() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%s %s ", rr.Domain.String(), TypeString(rr.Type))
	if rr.Class != IN {
		fmt.Fprintf(buf, "%s ", ClassString(rr.Class))
	}
	rr.Rdata.PrintTo(buf)
	fmt.Fprintf(buf, " %s", ttlStr(rr.TTL))

	return buf.String()
}

// Digest returns a one line digest of the rr record
func (rr *RR) Digest() string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%s %s ", rr.Domain.String(), TypeString(rr.Type))
	if rr.Class != IN {
		fmt.Fprintf(buf, "%s ", ClassString(rr.Class))
	}
	rr.Rdata.PrintTo(buf)

	return buf.String()
}
