package dig8

import (
	"bytes"
	"fmt"
)

func ttlStr(t uint32) string {
	if t == 0 {
		return "0"
	}

	buf := new(bytes.Buffer)
	second := t % 60
	minute := t / 60 % 60
	hour := t / 3600 % 24
	day := t / 3600 / 24
	if day > 0 {
		fmt.Fprintf(buf, "%dd", day)
	}
	if hour > 0 {
		fmt.Fprintf(buf, "%dh", hour)
	}
	if minute > 0 {
		fmt.Fprintf(buf, "%dm", minute)
	}
	if second > 0 {
		fmt.Fprintf(buf, "%ds", second)
	}

	return buf.String()
}
