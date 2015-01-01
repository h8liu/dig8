package dig8

// Respond is a crawler job feedback packet
type Respond struct {
	Name    string
	Crawled int
	Total   int
	Done    bool
	Error   string
}
