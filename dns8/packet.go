package dns8

import (
	"bytes"
	"errors"
	"math/rand"
)

// Packet is an DNS packet
type Packet struct {
	Bytes []byte

	ID        uint16
	Flag      uint16
	Question  *Question
	Answer    Section
	Authority Section
	Addition  Section
}

// Rcode returns the rcode of the packet
func (p *Packet) Rcode() uint16 {
	return p.Flag & RcodeMask
}

func randomID() uint16 { return uint16(rand.Uint32()) }

// Unpack unpacks a packet
func Unpack(p []byte) (*Packet, error) {
	m := new(Packet)
	m.Bytes = p
	m.Question = new(Question)

	e := m.unpack()

	return m, e
}

func (p *Packet) unpack() error {
	if p.Bytes == nil {
		return errors.New("nil packet")
	}

	in := bytes.NewReader(p.Bytes)

	if e := p.unpackHeader(in); e != nil {
		return e
	} else if e := p.Question.unpack(in, p.Bytes); e != nil {
		return e
	} else if e := p.Answer.unpack(in, p.Bytes); e != nil {
		return e
	}

	if p.Flag&FlagTC != 0 {
		p.Authority = p.Authority[0:0]
		p.Addition = p.Addition[0:0]
		return nil
	}

	if e := p.Authority.unpack(in, p.Bytes); e != nil {
		return e
	} else if e := p.Addition.unpack(in, p.Bytes); e != nil {
		return e
	}

	return nil
}

func (p *Packet) unpackHeader(in *bytes.Reader) error {
	buf := make([]byte, 12)
	if _, e := in.Read(buf); e != nil {
		return e
	}

	p.ID = enc.Uint16(buf[0:2])
	p.Flag = enc.Uint16(buf[2:4])
	if enc.Uint16(buf[4:6]) != 1 {
		return errors.New("not one question")
	}

	p.Answer = make([]*RR, enc.Uint16(buf[6:8]))
	p.Authority = make([]*RR, enc.Uint16(buf[8:10]))
	p.Addition = make([]*RR, enc.Uint16(buf[10:12]))

	return nil
}

func (p *Packet) packHeader(out *bytes.Buffer) {
	buf := make([]byte, 12)

	enc.PutUint16(buf[0:2], p.ID)
	enc.PutUint16(buf[2:4], p.Flag)
	enc.PutUint16(buf[4:6], 1) // always have one question
	enc.PutUint16(buf[6:8], p.Answer.LenU16())
	enc.PutUint16(buf[8:10], p.Authority.LenU16())
	enc.PutUint16(buf[10:12], p.Addition.LenU16())

	out.Write(buf)
}

// PackQuery packs a query.
func (p *Packet) PackQuery() []byte {
	out := new(bytes.Buffer)

	p.packHeader(out)
	p.Question.pack(out)

	p.Bytes = out.Bytes() // swap in
	return p.Bytes
}

// Qpack makes a query packet
func Qpack(d *Domain, t uint16) *Packet {
	return QpackID(d, t, randomID())
}

// QpackID makes a query pakcet with a particular id
func QpackID(d *Domain, t, id uint16) *Packet {
	m := new(Packet)

	if t == 0 {
		t = A
	}

	m.ID = id
	m.Flag = 0
	m.Question = &Question{d, t, IN}
	m.PackQuery()

	return m
}

// PrintTo prints the packet to a printer
func (p *Packet) PrintTo(prt *Printer) {
	prt.Printf("#%d %s", p.ID, flagString(p.Flag))
	prt.Printf("ques %v", p.Question)
	p.Answer.PrintNameTo(prt, "answ")
	p.Authority.PrintNameTo(prt, "auth")
	p.Addition.PrintNameTo(prt, "addi")
}

func (p *Packet) String() string {
	return PrintStr(p)
}

// SelectWith selects records with a selector
func (p *Packet) SelectWith(s Selector) []*RR {
	ret := make([]*RR, 0, 10)
	ret = SelectAppend(p.Answer, s, SecAnsw, ret)
	ret = SelectAppend(p.Authority, s, SecAuth, ret)
	ret = SelectAppend(p.Addition, s, SecAddi, ret)

	return ret
}

// SelectIPs selects A records for a domain.
func (p *Packet) SelectIPs(d *Domain) []*RR {
	return p.SelectWith(&SelectIP{d})
}

// SelectRedirects selects redirection related records
func (p *Packet) SelectRedirects(z, d *Domain) []*RR {
	return p.SelectWith(&SelectRedirect{z, d})
}

// SelectAnswers select answer records for a question
func (p *Packet) SelectAnswers(d *Domain, t uint16) []*RR {
	return p.SelectWith(&SelectAnswer{d, t})
}

// SelectRecords select records for of a particular type and
// domain
func (p *Packet) SelectRecords(d *Domain, t uint16) []*RR {
	return p.SelectWith(&SelectRecord{d, t})
}
