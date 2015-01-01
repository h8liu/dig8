package dig8

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strings"

	"lonnie.io/dig8/dns8"
)

// Worker is an RPC server that takes job requests and
// performs jobs.
type Worker struct {
	archive string // the archive path
}

func checkIdent(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '-' || r == '_' {
			continue
		}
		return false
	}

	return true
}

func checkName(name string) bool {
	p := strings.Index(name, ".")
	if p == -1 {
		return checkIdent(name)
	}

	folder := name[:p]
	file := name[p+1:]

	return checkIdent(folder) && checkIdent(file)
}

// Crawl is an RPC routine that accepts a request for crawling
// a crawler job.
func (w *Worker) Crawl(req *Request, err *string) error {
	if !checkName(req.Name) {
		*err = "bad job name"
		return nil
	}

	// convert to dns8.Domain list
	var doms []*dns8.Domain
	for _, d := range req.Domains {
		dom, e := dns8.ParseDomain(d)
		if e != nil {
			*err = e.Error()
			return nil
		}
		doms = append(doms, dom)
	}

	// ready to run now
	j := newJob(req.Name, doms, req.Callback)
	j.archive = w.archive
	go j.run()

	*err = "" // no error
	return nil
}

// WorkerServe runs a slave on a archive path on this machine.
func WorkerServe(archivePath string) {
	w := &Worker{
		archive: archivePath,
	}

	s := rpc.NewServer()
	e := s.RegisterName("Worker", w)
	if e != nil {
		log.Fatal(e)
	}
	rpc.HandleHTTP()

	addr := ":5353"
	log.Printf("listening on: %q\n", addr)

	conn, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatal("listen error:", e)
	}

	for {
		e = http.Serve(conn, s)
		if e != nil {
			log.Fatal("worker serve error:", e)
		}
	}
}
