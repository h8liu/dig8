package dig8

import (
	"bytes"
	"errors"
	"math/rand"
)

// Packet is an DNS packet
type Packet struct {
	Bytes []byte

	Id        uint16
	Flag      uint16
	Question  *Question
	Answer    Section
	Authority Section
	Addition  Section
}

// Rcode returns the rcode of the packet
func (s *Packet) Rcode() uint16 {
	return s.Flag & RcodeMask
}

func randomId() uint16 { return uint16(rand.Uint32()) }

// Unpack unpacks a packet
func Unpack(p []byte) (*Packet, error) {
	m := new(Packet)
	m.Bytes = p
	m.Question = new(Question)

	e := m.unpack()

	return m, e
}

func (s *Packet) unpack() error {
	if s.Bytes == nil {
		return errors.New("nil packet")
	}

	in := bytes.NewReader(s.Bytes)

	if e := s.unpackHeader(in); e != nil {
		return e
	} else if e := s.Question.unpack(in, s.Bytes); e != nil {
		return e
	} else if e := s.Answer.unpack(in, s.Bytes); e != nil {
		return e
	}

	if s.Flag&FlagTC != 0 {
		s.Authority = s.Authority[0:0]
		s.Addition = s.Addition[0:0]
		return nil
	}

	if e := s.Authority.unpack(in, s.Bytes); e != nil {
		return e
	} else if e := s.Addition.unpack(in, s.Bytes); e != nil {
		return e
	}

	return nil
}

func (s *Packet) unpackHeader(in *bytes.Reader) error {
	buf := make([]byte, 12)
	if _, e := in.Read(buf); e != nil {
		return e
	}

	s.Id = enc.Uint16(buf[0:2])
	s.Flag = enc.Uint16(buf[2:4])
	if enc.Uint16(buf[4:6]) != 1 {
		return errors.New("not one question")
	}

	s.Answer = make([]*RR, enc.Uint16(buf[6:8]))
	s.Authority = make([]*RR, enc.Uint16(buf[8:10]))
	s.Addition = make([]*RR, enc.Uint16(buf[10:12]))

	return nil
}

func (s *Packet) packHeader(out *bytes.Buffer) {
	buf := make([]byte, 12)

	enc.PutUint16(buf[0:2], s.Id)
	enc.PutUint16(buf[2:4], s.Flag)
	enc.PutUint16(buf[4:6], 1) // always have one question
	enc.PutUint16(buf[6:8], s.Answer.LenU16())
	enc.PutUint16(buf[8:10], s.Authority.LenU16())
	enc.PutUint16(buf[10:12], s.Addition.LenU16())

	out.Write(buf)
}

func (s *Packet) PackQuery() []byte {
	out := new(bytes.Buffer)

	s.packHeader(out)
	s.Question.pack(out)

	s.Bytes = out.Bytes() // swap in
	return s.Bytes
}

// Q makes a query packet
func Q(d *Domain, t uint16) *Packet {
	return Qid(d, t, randomId())
}

// Qid makes a query pakcet with a particular id
func Qid(d *Domain, t, id uint16) *Packet {
	m := new(Packet)

	if t == 0 {
		t = A
	}

	m.Id = id
	m.Flag = 0
	m.Question = &Question{d, t, IN}
	m.PackQuery()

	return m
}

// PrintTo prints the packet to a printer
func (s *Packet) PrintTo(p *Printer) {
	p.Printf("#%d %s", s.Id, flagString(s.Flag))
	p.Printf("ques %v", s.Question)
	s.Answer.PrintNameTo(p, "answ")
	s.Authority.PrintNameTo(p, "auth")
	s.Addition.PrintNameTo(p, "addi")
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
func (s *Packet) SelectIPs(d *Domain) []*RR {
	return s.SelectWith(&SelectIP{d})
}

// SelectRedirects selects redirection related records
func (s *Packet) SelectRedirects(z, d *Domain) []*RR {
	return s.SelectWith(&SelectRedirect{z, d})
}

// SelectAnswers select answer records for a question
func (s *Packet) SelectAnswers(d *Domain, t uint16) []*RR {
	return s.SelectWith(&SelectAnswer{d, t})
}

// SelectRecords select records for of a particular type and
// domain
func (s *Packet) SelectRecords(d *Domain, t uint16) []*RR {
	return s.SelectWith(&SelectRecord{d, t})
}
