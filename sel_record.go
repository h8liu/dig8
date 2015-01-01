package dig8

// SelectRecord selects records with a particular domain and type
type SelectRecord struct {
	Domain *Domain
	Type   uint16
}

// Select checks if the record is of the domain and type.
func (s *SelectRecord) Select(rr *RR, _ int) bool {
	return rr.Domain.Equal(s.Domain) && s.Type == rr.Type
}

var _ Selector = new(SelectRecord)
