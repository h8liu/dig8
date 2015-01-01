package dns8

// Leaf is a leaf node in a query tree
type Leaf struct {
	Attempts []*Exchange
}

var _ Node = new(Leaf)

// IsLeaf returns true.
func (lf *Leaf) IsLeaf() bool { return true }

func newLeaf(retry int) *Leaf {
	ret := new(Leaf)
	ret.Attempts = make([]*Exchange, 0, retry)
	return ret
}

func (lf *Leaf) add(e *Exchange) {
	lf.Attempts = append(lf.Attempts, e)
}

// Last returns the last attempt
func (lf *Leaf) Last() *Exchange {
	n := len(lf.Attempts)
	if n == 0 {
		return nil
	}
	return lf.Attempts[n-1]
}
