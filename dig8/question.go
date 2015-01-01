package dig8

import (
	"bytes"
	"fmt"
)

// Question is a DNS query question
type Question struct {
	Domain *Domain
	Type   uint16
	Class  uint16
}

func (q *Question) packFlags(out *bytes.Buffer) {
	buf := make([]byte, 4)
	enc.PutUint16(buf[0:2], q.Type)
	enc.PutUint16(buf[2:4], q.Class)
	out.Write(buf)
}

func (q *Question) pack(out *bytes.Buffer) {
	q.Domain.Pack(out)
	q.packFlags(out)
}

func (q *Question) unpack(in *bytes.Reader, p []byte) error {
	d, e := UnpackDomain(in, p)
	if e != nil {
		return e
	}
	q.Domain = d

	return q.unpackFlags(in)
}

func (q *Question) unpackFlags(in *bytes.Reader) error {
	buf := make([]byte, 4)
	if _, e := in.Read(buf); e != nil {
		return e
	}

	q.Type = enc.Uint16(buf[0:2])
	q.Class = enc.Uint16(buf[2:4])

	return nil
}

func (q *Question) String() string {
	ret := fmt.Sprintf("%s %s", q.Domain.String(), TypeString(q.Type))
	if q.Class != IN {
		ret += fmt.Sprintf(" %s", ClassString(q.Class))
	}
	return ret
}
