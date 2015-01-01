package dns8

import (
	"bytes"
	"fmt"
)

// RdTxt is a text
type RdTxt string

// UnpackRdTxt unpacks TXT record
func UnpackRdTxt(in *bytes.Reader, n uint16) (RdTxt, error) {
	buf := make([]byte, n)
	_, e := in.Read(buf)
	if e != nil {
		return "", e
	}
	return RdTxt(string(buf)), nil
}

// PrintTo prints it out
func (d RdTxt) PrintTo(out *bytes.Buffer) {
	fmt.Fprintf(out, "%#v", string(d))
}

// Pack packs it up
func (d RdTxt) Pack() []byte {
	return []byte(d)
}
