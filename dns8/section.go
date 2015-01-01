package dns8

import (
	"bytes"
)

// Section is a record section
type Section []*RR

// LenU16 returns the length of the section in uint16
func (s Section) LenU16() uint16 {
	if s == nil {
		return 0
	}

	if len(s) > 0xffff {
		panic("too many rrs")
	}

	return uint16(len(s))
}

// unpack unpacks the entire section
func (s Section) unpack(in *bytes.Reader, p []byte) error {
	var e error
	for i := range s {
		s[i], e = unpackRR(in, p)
		if e != nil {
			return e
		}
	}

	return nil
}

// PrintTo prints the section to a printer.
func (s Section) PrintTo(p *Printer) {
	for _, rr := range s {
		p.Print(rr)
	}
}

// PrintNameTo prints the section with a name.
func (s Section) PrintNameTo(p *Printer, name string) {
	if s == nil {
		return
	}
	if len(s) == 0 {
		return
	}

	if len(s) == 1 {
		p.Printf("%s %v", name, s[0])
	} else {
		p.Printf("%s {", name)
		p.ShiftIn()
		s.PrintTo(p)
		p.ShiftOut("}")
	}
}
