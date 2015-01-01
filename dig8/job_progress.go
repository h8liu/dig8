package dig8

// JobProgress is a crawler job feedback packet
type JobProgress struct {
	Name    string
	Crawled int
	Total   int
	Done    bool
	Error   string
}
