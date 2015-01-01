package dns8

// SelectIP selects A records for a particular domain
type SelectIP struct{ Domain *Domain }

// Select checks if the records is an A record for the domain.
func (s *SelectIP) Select(rr *RR, _ int) bool {
	return rr.Type == A && rr.Domain.Equal(s.Domain)
}

var _ Selector = new(SelectIP)
