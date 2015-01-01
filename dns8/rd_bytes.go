package dns8

import (
	"bytes"
	"fmt"
)

// RdBytes is just an array of bytes
type RdBytes []byte

var _ Rdata = RdBytes(nil)

// Pack packs the bytes.
func (bs RdBytes) Pack() []byte {
	ret := make([]byte, len(bs))
	copy(ret, bs)
	return ret
}

// UnpackRdBytes unpacks the rdata as bytes.
func UnpackRdBytes(in *bytes.Reader, n uint16) (RdBytes, error) {
	ret := make([]byte, n)
	if _, e := in.Read([]byte(ret)); e != nil {
		return nil, e
	}

	return RdBytes(ret), nil
}

// PrintTo prints it out
func (bs RdBytes) PrintTo(out *bytes.Buffer) {
	fmt.Fprintf(out, "[")
	for i, b := range bs {
		if i > 0 && i%4 == 0 {
			fmt.Fprintf(out, " ")
		}
		fmt.Fprintf(out, "%02x", b)
	}
	fmt.Fprintf(out, "]")
}
