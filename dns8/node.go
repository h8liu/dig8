package dns8

// Node is a node in a query tree.
type Node interface {
	// IsLeaf checks if the node is a leaf
	// If it is not a leaf, it must be a branch.
	IsLeaf() bool
}
