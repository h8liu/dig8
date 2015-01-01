package dns8

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// RdSoa is an SOA rdata
type RdSoa struct {
	Mname                  []string
	Rname                  []string
	Serial, Refresh        uint32
	Retry, Expire, Minimum uint32
}

// PrintTo prints the thing to output
func (d *RdSoa) PrintTo(out *bytes.Buffer) {
	fmt.Fprintf(out, "%v/%v serial=%d refresh=%d retry=%d exp=%d min=%d",
		strings.Join(d.Mname, "."),
		strings.Join(d.Rname, "."),
		d.Serial, d.Refresh, d.Retry, d.Expire, d.Minimum)
}

// UnpackRdSoa unpacks the thing
func UnpackRdSoa(in *bytes.Reader, n uint16, p []byte) (*RdSoa, error) {
	if n <= 22 {
		return nil, fmt.Errorf("soa with %d bytes", n)
	}

	ret := new(RdSoa)
	was := in.Len()
	labels, e := UnpackLabels(in, p)
	if e != nil {
		return nil, e
	}
	ret.Mname = labels

	labels, e = UnpackLabels(in, p)
	if e != nil {
		return nil, e
	}
	ret.Rname = labels

	now := in.Len()
	if was-now+20 != int(n) {
		return nil, errors.New("invalid soa field length")
	}

	buf := make([]byte, 20)
	_, e = in.Read(buf)
	if e != nil {
		return nil, e
	}
	ret.Serial = enc.Uint32(buf[0:4])
	ret.Refresh = enc.Uint32(buf[4:8])
	ret.Retry = enc.Uint32(buf[8:12])
	ret.Expire = enc.Uint32(buf[12:16])
	ret.Minimum = enc.Uint32(buf[16:20])

	return ret, nil
}

// Pack packs the thing.
func (d *RdSoa) Pack() []byte {
	buf := new(bytes.Buffer)
	PackLabels(buf, d.Mname)
	PackLabels(buf, d.Rname)

	b := make([]byte, 20)
	enc.PutUint32(b[0:4], d.Serial)
	enc.PutUint32(b[4:8], d.Refresh)
	enc.PutUint32(b[8:12], d.Retry)
	enc.PutUint32(b[12:16], d.Expire)
	enc.PutUint32(b[16:20], d.Minimum)
	buf.Write(b)

	return buf.Bytes()

}
