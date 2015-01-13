package dcrl

// RPCServer wraps the sever with the exported RPC interfaces
type RPCServer Server

// ClaimJob wraps the server's ClaimJob
func (s *RPCServer) ClaimJob(worker string, j *JobDesc) error {
	c := make(chan error)
	(*Server)(s).requests <- &request{
		typ:   "claimJob",
		data:  worker,
		reply: j,
		c:     c,
	}

	return <-c
}

// Progress wraps the server's Progress
func (s *RPCServer) Progress(p *Progress, okay *bool) error {
	c := make(chan error)
	(*Server)(s).requests <- &request{
		typ:   "progress",
		data:  p,
		reply: okay,
		c:     c,
	}

	return <-c
}

// NewJob creates a new job with a particular tag
func (s *RPCServer) NewJob(j *NewJobDesc, name *string) error {
	c := make(chan error)
	(*Server)(s).requests <- &request{
		typ:   "newJob",
		data:  j,
		reply: name,
		c:     c,
	}

	return <-c
}
