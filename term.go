package dig8

import (
	"os"
)

// Term is a query terminal that uses a client
// and builds query trees.
type Term struct {
	client *Client
	done   int

	*TermConfig
}

// NewTerm creates a new query terminal
func NewTerm(c *Client) *Term {
	ret := new(Term)
	ret.client = c
	ret.PrintFlag = PrintReply
	ret.Retry = 3

	return ret
}

// T builds a query tree in the terminal
func (tm *Term) T(t Task) (*Branch, error) {
	ret, e := newCursor(tm.TermConfig, tm.client).T(t)
	tm.done++

	if e == nil {
		p := NewPrinter(tm.Out)
		t.PrintTo(p)
	}

	return ret, e
}

// Q builds a query tree leaf in the terminal
func (tm *Term) Q(q *Query) (*Leaf, error) {
	ret, e := newCursor(tm.TermConfig, tm.client).Q(q)
	tm.done++

	return ret, e
}

// Count returns the number of queries done.
func (tm *Term) Count() int { return tm.done }

var stdTerm *Term

func makeStdTerm() *Term {
	if stdTerm == nil {
		c, e := NewClient()
		if e != nil {
			panic(e)
		}

		stdTerm = NewTerm(c)
		stdTerm.Log = os.Stdout
		stdTerm.Out = os.Stdout
	}

	return stdTerm
}

// DoTask performs a task with the default term
// and returns a query tree.
func DoTask(t Task) *Branch {
	ret, e := makeStdTerm().T(t)
	if e != nil {
		panic(e)
	}
	return ret
}

// DoQuery performs a query with the default term
// and returns a query tree leaf.
func DoQuery(q *Query) *Leaf {
	ret, e := makeStdTerm().Q(q)
	if e != nil {
		panic(e)
	}
	return ret
}
