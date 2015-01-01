package dns8

// SelectAnswer selects answer records.
type SelectAnswer struct {
	Domain *Domain
	Type   uint16
}

// Select checks if the records is an answer.
func (s *SelectAnswer) Select(rr *RR, _ int) bool {
	if !rr.Domain.Equal(s.Domain) {
		return false
	}
	return s.Type == rr.Type || (s.Type == A && rr.Type == CNAME)
}

var _ Selector = new(SelectAnswer)
