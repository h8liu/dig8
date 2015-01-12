package dcrl

import (
	"database/sql"
)

// Server is a stateless server where every call is done
// by a transaction on the underlying sqlite database
type Server struct {
	Addr     string
	JobsDB   string
	JobsPath string

	requests chan *request
	db       *sql.DB
}

type request struct {
	typ   string
	data  interface{}
	reply interface{}
	c     chan error
}

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
func (s *RPCServer) Progress(p *Progress, hit *bool) error {
	c := make(chan error)
	(*Server)(s).requests <- &request{
		typ:   "progress",
		data:  p,
		reply: hit,
		c:     c,
	}

	return <-c
}

// Serve launches the server
func (s *Server) Serve() error {
	var e error
	s.db, e = sql.Open("sqlite3", s.JobsDB)
	if e != nil {
		return e
	}
	defer s.db.Close()

	s.requests = make(chan *request, 64)

	return serve(s)
}
