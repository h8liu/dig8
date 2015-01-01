package dig8

// Request is a crawling request for a list of domains.
type JobRequest struct {
	Name     string
	Domains  []string
	Callback string
}
