package dcrl

// JobDesc is a job descriptor
type JobDesc struct {
	Name    string
	Archive string
	Domains []string
}

// NewJobDesc describes a new job request
type NewJobDesc struct {
	Tag     string
	Archive string
	Domains []string
}
