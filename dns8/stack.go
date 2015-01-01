package dns8

// stack is a query stack for developing a query tree
type stack struct {
	s []*Branch
}

func newStack() *stack {
	ret := new(stack)
	ret.s = make([]*Branch, 0, 20)
	return ret
}

func (s *stack) Push(b *Branch) {
	s.s = append(s.s, b)
}

func (s *stack) Pop() *Branch {
	n := len(s.s)
	if n == 0 {
		return nil
	}
	ret := s.s[n-1]
	s.s = s.s[:n-1]
	return ret
}

func (s *stack) Len() int {
	return len(s.s)
}

func (s *stack) Top() *Branch {
	n := len(s.s)
	if n == 0 {
		return nil
	}
	return s.s[n-1]
}

func (s *stack) TopAdd(n Node) {
	t := s.Top()
	if t != nil {
		t.add(n)
	}
}
