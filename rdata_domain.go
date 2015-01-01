package dig8

import (
	"bytes"
	"errors"
	"fmt"
)

// RdDomain is rdata of domain
type RdDomain Domain

// PrintTo prints the rdata to out
func (d *RdDomain) PrintTo(out *bytes.Buffer) {
	fmt.Fprint(out, (*Domain)(d))
}

// Pack packs the domain
func (d *RdDomain) Pack() []byte {
	buf := new(bytes.Buffer)
	(*Domain)(d).Pack(buf)
	return buf.Bytes()
}

// UnpackRdDomain unpacks the domain
func UnpackRdDomain(in *bytes.Reader, n uint16, p []byte) (*RdDomain, error) {
	if n == 0 {
		return nil, errors.New("zero domain len")
	}

	was := in.Len()
	d, e := UnpackDomain(in, p)
	now := in.Len()
	if was-now != int(n) {
		return nil, fmt.Errorf("domain len expect %d, got %d", n, was-now)
	}

	return (*RdDomain)(d), e
}

// RdToDomain converts a domain rddata to Domain
func RdToDomain(r Rdata) *Domain {
	return (*Domain)(r.(*RdDomain))
}
