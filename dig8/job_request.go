package dig8

// JobRequest is a crawling request for a list of domains.
type JobRequest struct {
	Name     string
	Domains  []string
	Archive  string
	Callback string
}
