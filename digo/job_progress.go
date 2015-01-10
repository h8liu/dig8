package digo

// JobProgress is a crawler job feedback packet
type JobProgress struct {
	Name    string
	Crawled int
	Total   int
	Done    bool
	Error   string
}

func (p *JobProgress) done() bool {
	return p.Done || p.Error != ""
}
