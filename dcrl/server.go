package dcrl

import (
	"database/sql"
)

// Server is a stateless server where every call is done
// by a transaction on the underlying sqlite database
type Server struct {
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
