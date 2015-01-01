package dns8

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// RdMx is a MX rdata
type RdMx struct {
	Priority uint16
	Domain   []string
}

// PrintTo prints the thing
func (d *RdMx) PrintTo(out *bytes.Buffer) {
	fmt.Fprintf(out, "%s/%d", strings.Join(d.Domain, "."),
		d.Priority)
}

func unpackMxLabels(in *bytes.Reader, n uint16, p []byte) ([]string, error) {
	if n == 0 {
		return nil, errors.New("zero labels len")
	}

	was := in.Len()
	d, e := UnpackLabels(in, p)
	now := in.Len()
	if was-now != int(n) {
		return nil, fmt.Errorf("domain length expect %d, actual %d",
			n, was-now)
	}

	return d, e
}

// UnpackRdMx unpacks the thing
func UnpackRdMx(in *bytes.Reader, n uint16, p []byte) (*RdMx, error) {
	if n <= 2 {
		return nil, fmt.Errorf("mx with %d bytes", n)
	}

	buf := make([]byte, 2)
	_, e := in.Read(buf)
	if e != nil {
		return nil, e
	}

	ret := new(RdMx)
	ret.Priority = enc.Uint16(buf)
	labels, e := unpackMxLabels(in, n-2, p)
	if e != nil {
		return nil, e
	}
	ret.Domain = labels

	return ret, nil
}

// Pack packs the thing
func (d *RdMx) Pack() []byte {
	buf := new(bytes.Buffer)
	b := make([]byte, 2)
	enc.PutUint16(b, d.Priority)
	buf.Write(b)
	PackLabels(buf, d.Domain)
	return buf.Bytes()
}
