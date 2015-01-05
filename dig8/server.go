package dig8

import (
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"strings"
	"time"

	"lonnie.io/dig8/dns8"

	_ "github.com/mattn/go-sqlite3" // sqlite3 support
)

func qne(sql string, e error) {
	if e != nil {
		log.Printf("sql: %s\n", sql)
		ne(e)
	}
}

// InitDB creates the database schemes
func InitDB(dbPath string) {
	db, err := sql.Open("sqlite3", dbPath)
	ne(err)

	q := func(sql string) {
		_, e := db.Exec(sql)
		qne(sql, e)
	}

	q(`create table jobs (
		name text not null primary key,
		state int not null,
		total int not null,
		crawled int not null,
		sample text not null,
		salt text not null,
		worker text not null default "",
		err text not null default "",
		birth text not null default "",
		death text not null default ""
	);`)

	ne(db.Close())
}

// Server is a server working on a database.
type Server struct {
	db     *sql.DB
	cbAddr string
}

// NewServer creates a new server working on a database.
func NewServer(dbPath string, cbAddr string) (*Server, error) {
	ret := new(Server)

	db, e := sql.Open("sqlite3", dbPath)
	if e != nil {
		return nil, e
	}

	ret.db = db
	ret.cbAddr = cbAddr

	return ret, nil
}

// db.Exec wrapper
func (s *Server) q(sql string, args ...interface{}) sql.Result {
	res, e := s.db.Exec(sql, args...)
	qne(sql, e)
	return res
}

// db.Query wrapper
func (s *Server) qs(sql string, args ...interface{}) *sql.Rows {
	rows, e := s.db.Query(sql, args...)
	qne(sql, e)
	return rows
}

// State is the state code of a job
type State int

// Job states
const (
	Registered State = iota
	Created
	Crawling
	Done
	Errored
	Archived
)

var archiveDelay = time.Second * 10

func (s *Server) archive(name string) {
	go func(name string) {
		time.Sleep(archiveDelay)
		s.q(`update jobs set state=? where name=?`,
			int(Archived), name,
		)
	}(name)
}

func jsonEncode(i interface{}) string {
	bs, e := json.MarshalIndent(i, "", "    ")
	ne(e)
	return string(bs)
}

// RpcServer for rpc
type RPCServer struct {
	s *Server
}

// CallbackServer for callback
type CallbackServer struct {
	s *Server
}

// Progress wraps the Progress function of the server
func (s *CallbackServer) Progress(p *JobProgress, err *string) error {
	return s.s.Progress(p, err)
}

// NewJob wraps the NewJob function of the server
func (s *RPCServer) NewJob(j *NewJob, err *string) error {
	return s.s.NewJob(j, err)
}

// RPC returns an RPCServer
func (s *Server) RPC() *RPCServer { return &RPCServer{s} }

// Callback returns a callback server
func (s *Server) Callback() *CallbackServer { return &CallbackServer{s} }

// Progress updates the progress.
func (s *Server) Progress(p *JobProgress, err *string) error {
	log.Println("Progres: ", jsonEncode(p))

	state := Crawling
	if p.Error != "" {
		state = Errored
	} else if p.Done {
		state = Done
		s.archive(p.Name)
	}

	s.q(`update jobs set state=?, crawled=?, err=? where name=?;`,
		int(state), p.Crawled, p.Error, p.Name,
	)
	*err = ""

	return nil
}

// NewJob is a new job request
type NewJob struct {
	Tag     string
	Domains []string
}

const encodeHex = "0123456789abcdefghijklmnopqrstuv"

var base32enc = base32.NewEncoding(encodeHex)

func (s *Server) errorJob(name string, e error) {
	log.Println("job %q error: %s", name, e)
	s.q(`update jobs set state=? err=? where name=?`,
		int(Errored), e.Error(), name,
	)
}

func (s *Server) createJob(doms []string, name string) {
	f, e := os.Create(name)
	if e != nil {
		s.errorJob(name, e)
		return
	}

	for _, d := range doms {
		_, e := dns8.ParseDomain(d)
		if e != nil {
			s.errorJob(name, e)
			f.Close()
			return
		}
	}

	_, e = io.WriteString(f, strings.Join(doms, "\n"))
	if e != nil {
		s.errorJob(name, e)
		f.Close()
		return
	}

	e = f.Close()
	if e != nil {
		s.errorJob(name, e)
		return
	}

	s.q(`update jobs set state=? where name=?`, int(Created), name)
}

func (s *Server) pickWorker() string {
	return "localhost:5353"
}

func (s *Server) startJob(name string) {
	bs, e := ioutil.ReadFile(name)
	if e != nil {
		s.errorJob(name, e)
		return
	}

	doms := strings.Split(string(bs), "\n")
	// total := len(doms)

	worker := s.pickWorker()
	s.q(`update job set worker=?, state=? where name=?`,
		worker, int(Crawling), name,
	)

	c, e := rpc.DialHTTP("tcp", worker)
	if e != nil {
		s.errorJob(name, e)
		return
	}

	req := new(JobRequest)
	req.Name = name
	req.Domains = doms
	req.Callback = s.cbAddr

	var err string
	e = c.Call("Worker.Crawl", req, &err)
	if e != nil {
		s.errorJob(name, e)
		return
	}
	if err != "" {
		s.errorJob(name, errors.New(err))
		return
	}
}

// NewJob creates a new job
func (s *Server) NewJob(j *NewJob, err *string) error {
	log.Println("NewJob: ", jsonEncode(j))

	nsample := len(j.Domains)
	if nsample > 20 {
		nsample = 20
	}
	doms := j.Domains[:nsample]
	sample := strings.Join(doms, "\n")
	createTime := time.Now().String()

	var salt [32]byte
	var name string

	var succ bool
	for i := 0; i < 100; i++ {
		// generate salt
		_, e := rand.Read(salt[:])
		ne(e)
		saltStr := base32enc.EncodeToString(salt[:])

		// generate name based on salted hash of the sample
		hash := sha1.New()
		hash.Write(salt[:])
		io.WriteString(hash, sample)
		name = j.Tag + "." + base32enc.EncodeToString(hash.Sum(nil))[:6]

		res := s.q(`insert or ignore into jobs 
			(name, state, total, crawled, sample, salt, birth)
			values
			(?, ?, ?, ?, ?, ?, ?)`,
			name, int(Registered), len(j.Domains), 0,
			sample, saltStr, createTime,
		)

		rows, e := res.RowsAffected()
		ne(e)
		if rows > 0 {
			succ = true
			break
		}
	}

	if !succ {
		*err = "creating job failed, tried 100 times"
		return nil
	}

	s.createJob(j.Domains, name)
	log.Println("job %q created", name)

	go s.startJob(name)

	*err = ""
	return nil
}

// ServeHTTP wraps the server as a http handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/")

	switch name {
	case "list":
		panic("todo")
	default:
		w.WriteHeader(404)
	}
}
