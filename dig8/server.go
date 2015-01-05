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
	mrand "math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"
	"sync"
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
		archive text not null,
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

	workersLock sync.Mutex
	workers     map[string]int
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
	ret.workers = make(map[string]int)

	return ret, nil
}

// ServeRPC launches the RPC server
func (s *Server) ServeRPC(rpcAddr string) {
	log.Printf("servering rpc on %s", rpcAddr)
	rs := rpc.NewServer()
	e := rs.RegisterName("Server", s.RPC())
	ne(e)

	c, e := net.Listen("tcp", rpcAddr)
	ne(e)

	for {
		e = http.Serve(c, rs)
		if e != nil {
			log.Print(e)
		}
	}
}

// ServeCallback launches the callback RPC server
func (s *Server) ServeCallback() {
	log.Printf("servering callback on %s", s.cbAddr)
	cb := rpc.NewServer()
	e := cb.RegisterName("Cb", s.Callback())
	ne(e)
	c, e := net.Listen("tcp", s.cbAddr)
	ne(e)

	for {
		e = http.Serve(c, cb)
		if e != nil {
			log.Print(e)
		}
	}
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

// RPCServer for rpc
type RPCServer struct{ s *Server }

// CallbackServer for callback
type CallbackServer struct{ s *Server }

// Progress wraps the Progress function of the server
func (s *CallbackServer) Progress(p *JobProgress, err *string) error {
	return s.s.Progress(p, err)
}

// Heartbeat wraps the worker Heartbeat function of the server
func (s *CallbackServer) Heartbeat(ws *WorkerState, err *string) error {
	return s.s.Heartbeat(ws, err)
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

// Heartbeat reports the worker state
func (s *Server) Heartbeat(ws *WorkerState, err *string) error {
	s.workersLock.Lock()
	s.workers[ws.Worker] = ws.State
	s.workersLock.Unlock()

	*err = ""
	return nil
}

// NewJob is a new job request
type NewJob struct {
	Tag     string
	Archive string // archive saving position
	Domains []string
}

const encodeHex = "0123456789abcdefghijklmnopqrstuv"

var base32enc = base32.NewEncoding(encodeHex)

func (s *Server) errorJob(name string, e error) bool {
	if e == nil {
		return false
	}
	log.Printf("job %q error: %s", name, e)
	s.q(`update jobs set state=? err=? where name=?`,
		int(Errored), e.Error(), name,
	)
	return true
}

func (s *Server) createJob(doms []string, name string) {
	f, e := os.Create(name)
	if s.errorJob(name, e) {
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

	s.errorJob(name, f.Close())

	s.q(`update jobs set state=? where name=?`, int(Created), name)
}

func (s *Server) pickWorker() string {
	var workers []string

	s.workersLock.Lock()
	defer s.workersLock.Unlock()

	for w, state := range s.workers {
		if state == workerIdle {
			workers = append(workers, w)
		}
	}

	if len(workers) == 0 {
		return ""
	}

	ret := workers[mrand.Intn(len(workers))]
	s.workers[ret] = workerPending
	return ret
}

func (s *Server) startJob(name string) {
	bs, e := ioutil.ReadFile(name)
	if s.errorJob(name, e) {
		return
	}

	doms := strings.Split(string(bs), "\n")
	// total := len(doms)

	rows := s.qs("select archive from jobs where name=?", name)
	if !rows.Next() {
		log.Printf("row missing or error: %s", name)
		return
	}
	var archive string

	if s.errorJob(name, rows.Scan(&archive)) {
		return
	}
	if s.errorJob(name, rows.Err()) {
		return
	}
	rows.Close()

	worker := s.pickWorker()
	for worker == "" {
		time.Sleep(time.Minute) // TODO: this is way too ugly...
		worker = s.pickWorker()
	}

	s.q(`update jobs set worker=?, state=? where name=?`,
		worker, int(Crawling), name,
	)

	c, e := rpc.DialHTTP("tcp", worker)
	if s.errorJob(name, e) {
		return
	}

	req := new(JobRequest)
	req.Name = name
	req.Domains = doms
	req.Archive = archive
	req.Callback = s.cbAddr

	var err string
	e = c.Call("Worker.Crawl", req, &err)
	if s.errorJob(name, e) {
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
		name = base32enc.EncodeToString(hash.Sum(nil))[:6]
		if j.Tag != "" {
			name = j.Tag + "." + name
		}

		res := s.q(`insert or ignore into jobs 
			(name, archive, state, total, crawled, sample, salt, birth)
			values
			(?, ?, ?, ?, ?, ?, ?, ?)`,
			name, j.Archive, int(Registered), len(j.Domains), 0,
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
	log.Printf("job %q created", name)

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
