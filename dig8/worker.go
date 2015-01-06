package dig8

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strings"
	"time"

	"lonnie.io/dig8/dns8"
)

// Worker is an RPC server that takes job requests and
// performs jobs.
type Worker struct {
	archive    string // the archive path
	serverAddr string // for heartbeat
	workerAddr string // for heartbeat

	quotas chan bool
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

const nconcurrent = 3

func checkName(name string) bool {
	p := strings.Index(name, ".")
	if p == -1 {
		return checkIdent(name)
	}

	folder := name[:p]
	file := name[p+1:]

	return checkIdent(folder) && checkIdent(file)
}

func (w *Worker) heartbeat() {
	for {
		c, e := rpc.DialHTTP("tcp", w.serverAddr)
		if e != nil {
			time.Sleep(time.Minute)
			continue
		}

		state := workerIdle
		if len(w.quotas) == 0 {
			state = workerBusy
		}

		var err string
		e = c.Call("Cb.Heartbeat", &WorkerState{
			Worker: w.workerAddr,
			State:  state,
		}, &err)

		if e != nil {
			log.Print(e)
		} else if err != "" {
			log.Print(err)
		}

		e = c.Close()
		if e != nil {
			log.Print(e)
		}

		time.Sleep(time.Second * 5)
	}
}

// Crawl is an RPC routine that accepts a request for crawling
// a crawler job.
func (w *Worker) Crawl(req *JobRequest, err *string) error {
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
	if req.Archive == "" {
		j.archive = w.archive
	} else {
		j.archive = req.Archive
	}
	go func() {
		<-w.quotas
		j.run()
		<-j.jobDone
		w.quotas <- true
	}()

	*err = "" // no error
	return nil
}

// WorkerServe runs a slave on a archive path on this machine.
func WorkerServe(archivePath, serverAddr, workerAddr, listenAddr string) {
	w := &Worker{
		archive:    archivePath,
		serverAddr: serverAddr,
		workerAddr: workerAddr,
	}

	w.quotas = make(chan bool, nconcurrent)
	for i := 0; i < nconcurrent; i++ {
		w.quotas <- true
	}

	go w.heartbeat()

	s := rpc.NewServer()
	e := s.RegisterName("Worker", w)
	if e != nil {
		log.Fatal(e)
	}

	log.Printf("listening on: %q\n", listenAddr)

	conn, e := net.Listen("tcp", listenAddr)
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
