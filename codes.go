package dig8

import (
	"fmt"
)

// rdata type
const (
	A     = 1
	NS    = 2
	MD    = 3
	MF    = 4
	CNAME = 5
	SOA   = 6
	MB    = 7
	MG    = 8
	MR    = 9
	NULL  = 10
	WKS   = 11
	PTR   = 12
	HINFO = 13
	MINFO = 14
	MX    = 15
	TXT   = 16
	AAAA  = 28
)

// class code
const (
	IN = 1
	CS = 2
	CH = 3
	HS = 4
)

var (
	typeStrings = map[uint16]string{
		A:     "a",
		AAAA:  "aaaa",
		NS:    "ns",
		MX:    "mx",
		CNAME: "cname",
		TXT:   "txt",
		SOA:   "soa",
		NULL:  "null",
		PTR:   "ptr",
	}

	classStrings = map[uint16]string{
		IN: "in",
		CS: "cs",
		CH: "ch",
		HS: "hs",
	}
)

// TypeString returns the string of a type field
func TypeString(t uint16) string {
	s, found := typeStrings[t]
	if found {
		return s
	}
	return fmt.Sprintf("t%d", t)
}

// ClassString returns the string of a class field
func ClassString(c uint16) string {
	s, found := classStrings[c]
	if found {
		return s
	}
	return fmt.Sprintf("c%d", c)
}
