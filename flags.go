package dig8

import (
	"fmt"
	"strings"
)

// DNS flag field codes
const (
	FlagResponse = 0x1 << 15

	FlagRA = 0x1 << 7
	FlagRD = 0x1 << 8
	FlagTC = 0x1 << 9
	FlagAA = 0x1 << 10

	RcodeMask = 0xf
	OpMask    = 0x3 << 11
)

// DNS op field codes
const (
	OpQuery = iota << 11
	OpIquery
	OpStatus
)

// DNS rcodes
const (
	RcodeOkay = iota
	RcodeFormatError
	RcodeServerFail
	RcodeNameError
	RcodeNotImplement
	RcodeRefused
)

func flagString(flag uint16) string {
	var tags []string

	rcode := func(flag uint16) uint16 { return flag & RcodeMask }

	tag := func(b bool, s string) {
		if b {
			tags = append(tags, s)
		}
	}

	tag((flag&FlagResponse) == 0, "query")
	tag((flag&OpMask) == OpStatus, "status")
	tag((flag&OpMask) == OpIquery, "iquery")
	tag((flag&FlagAA) != 0, "auth")
	tag((flag&FlagTC) != 0, "trunc")
	tag((flag&FlagRD) != 0, "rec-desir")
	tag((flag&FlagRA) != 0, "rec-avail")

	c := rcode(flag)
	tag(c == RcodeFormatError, "fmt-err")
	tag(c == RcodeServerFail, "serv-fail")
	tag(c == RcodeNameError, "name-err")
	tag(c == RcodeNotImplement, "not-impl")
	tag(c == RcodeRefused, "refused")
	tag(c > RcodeRefused, fmt.Sprintf("rcode%d", c))

	return strings.Join(tags, " ")
}
