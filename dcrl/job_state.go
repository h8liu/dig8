package dcrl

// State is the state of a job
type State int

// The states of a job
const (
	Registered State = iota
	Created
	Crawling
	Errored
	Done
	Archived
)
