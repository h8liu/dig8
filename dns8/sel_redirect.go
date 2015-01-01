package dns8

// SelectRedirect selects records that are useful
// for server redirection
type SelectRedirect struct {
	Zone, Domain *Domain
}

// Select checks if the record is useful for
// server redirection
func (s *SelectRedirect) Select(rr *RR, _ int) bool {
	return rr.Type == NS && rr.Domain.IsChildOf(s.Zone) &&
		rr.Domain.IsZoneOf(s.Domain)
}

var _ Selector = new(SelectRedirect)
