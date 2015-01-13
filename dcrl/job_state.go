package dcrl

// State is the state of a job
type State int

// The states of a job
const (
	Created State = iota
	Crawling
	Errored
	Done
	Archived
)
