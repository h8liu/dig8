package dig8

import (
	"errors"
)

// cursor is a query cursor in a query terminal.
// It generates a query tree
type cursor struct {
	*Printer
	*TermConfig // conveniently inherits the term options
	*stack

	client *Client
	nquery int
	e      error
}

var _ Cursor = new(cursor)

func newCursor(cfg *TermConfig, c *Client) *cursor {
	ret := new(cursor)

	ret.TermConfig = cfg
	ret.stack = newStack()
	ret.Printer = NewPrinter(cfg.Log)
	ret.client = c

	return ret
}

// Query tree limits
const (
	MaxDepth = 30
	MaxQuery = 500
)

var (
	errTooDeep        = errors.New("too deep")
	errTooManyQueries = errors.New("too many queries")
)

// P returns the printer.
func (c *cursor) P() *Printer { return c.Printer }

// Error returns the cursor error, if any
func (c *cursor) E() error { return c.e }

// Q queries a query with the cursor.
func (c *cursor) Q(q *Query) (*Leaf, error) {
	if c.e != nil {
		return nil, c.e
	}

	if c.nquery >= MaxQuery {
		c.e = errTooManyQueries
		c.Printf("error %v", c.e)
		return nil, c.e
	}

	c.nquery++
	ret := c.q(q)
	c.TopAdd(ret)
	return ret, c.e
}

// T queries a task with the cursor
func (c *cursor) T(t Task) (*Branch, error) {
	if c.e != nil {
		return nil, c.e
	}

	if c.Len() >= MaxDepth {
		c.e = errTooDeep
		c.Printf("error %v", c.e)
		return nil, c.e
	}

	ret := newBranch(t)
	c.TopAdd(ret)
	c.Push(ret)

	t.Run(c)

	b := c.Pop()
	if b != ret || b.Task != t {
		panic("bug")
	}

	return ret, c.e
}

func (c *cursor) q(q *Query) *Leaf {
	qp := &QueryPrinter{
		Query:     q,
		Printer:   c.Printer,
		PrintFlag: c.PrintFlag,
	}

	ret := newLeaf(c.Retry)

	for i := 0; i < c.Retry; i++ {
		if i > 0 {
			c.Print("// retry")
		}
		answer := c.client.Query(qp)
		ret.add(answer)
		if answer.Timeout() {
			continue
		}
		break
	}

	return ret
}
