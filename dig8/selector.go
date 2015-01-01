package dig8

// Section flags
const (
	SecAnsw = 1 << iota // TODO: why use bits
	SecAuth
	SecAddi
)

// Selector is an interface for selecting records
type Selector interface {
	Select(rr *RR, section int) bool
}

// SelectAppend selects records from a section and appends
// it into ret, and returns the appended list.
// TODO: better interface?
func SelectAppend(s Section, sel Selector, section int, ret []*RR) []*RR {
	for _, rr := range s {
		if rr.Class != IN {
			continue
		}
		if sel.Select(rr, section) {
			ret = append(ret, rr)
		}
	}

	return ret
}
