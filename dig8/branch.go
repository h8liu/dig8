package dig8

// Branch is a branch in a query tree
type Branch struct {
	Task     // the task binded
	Children []Node
}

var _ Node = new(Branch)

func newBranch(t Task) *Branch {
	ret := new(Branch)
	ret.Task = t
	ret.Children = make([]Node, 0, 5)

	return ret
}

func (br *Branch) add(n Node) {
	if br == nil {
		return
	}
	br.Children = append(br.Children, n)
}

// IsLeaf returns false for branch
func (br *Branch) IsLeaf() bool { return false }

// Cursor defines the interface a task requires to build a query tree
type Cursor interface {
	P() *Printer
	E() error
	T(t Task) (*Branch, error)
	Q(q *Query) (*Leaf, error)
}

// Task is an executable node that builds a query tree
type Task interface {
	Printable
	Run(c Cursor)
}
