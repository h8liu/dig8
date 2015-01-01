package dns8

import (
	"bytes"
)

// PackRdata packs the rdata into the packet buffer.
func PackRdata(out *bytes.Buffer, rdata Rdata) {
	pack := rdata.Pack()
	n := len(pack)
	if n > 255 {
		panic("rdata too long")
	}

	buf := make([]byte, 2)
	enc.PutUint16(buf, uint16(n))
	out.Write(buf)
	out.Write(pack)
}
