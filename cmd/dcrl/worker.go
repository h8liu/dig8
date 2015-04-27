package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"path/filepath"
	"time"

	"lonnie.io/dig8/dcrl"
	"lonnie.io/dig8/dns8"
)

func workerName(i int) string {
	host, e := os.Hostname()
	ne(e)

	return fmt.Sprintf("%s-%d", host, i)
}

func parseDomains(doms []string) ([]*dns8.Domain, error) {
	ret := make([]*dns8.Domain, len(doms))
	for i, d := range doms {
		dom, e := dns8.ParseDomain(d)
		if e != nil {
			return nil, e
		}

		ret[i] = dom
	}

	return ret, nil
}

var backoff = time.Second

func work(addr string, i int, archive, logPath string) error {
	name := workerName(i)

	c, e := rpc.DialHTTP("tcp", addr)
	if e != nil {
		return e
	}

	defer c.Close()

	for {
		var job dcrl.JobDesc
		e = c.Call("Server.ClaimJob", name, &job)
		if e != nil {
			return e
		}

		if job.Name == "" {
			time.Sleep(backoff)
			continue
		}

		doms, e := parseDomains(job.Domains)
		if e != nil {
			log.Printf("[%s] error %s", job.Name, e)
			continue
		}

		arch := job.Archive
		if archive != "" {
			arch = filepath.Join(archive, job.Archive)
		}

		j := &dcrl.Job{
			Name:    job.Name,
			Archive: arch,
			Domains: doms,
			Log:     logPath,
			Progress: func(p *dcrl.Progress) error {
				var okay bool
				e = c.Call("Server.Progress", p, &okay)
				if e != nil {
					return e
				}
				if !okay {
					return fmt.Errorf("[%s] progress fail", job.Name)
				}
				return nil
			},
		}

		e = j.Do()
		if e != nil {
			return e
		}
	}
}

func workForever(addr string, i int, archive, logPath string) {
	for {
		e := work(addr, i, archive, logPath)
		if e != nil {
			log.Print(e)
		}
		time.Sleep(backoff)
	}
}

var (
	nworker  = flag.Int("nworker", 5, "concurrent worker")
	workaddr = flag.String("workaddr", "localhost:5300", "server address")
	archPath = flag.String("arch", "archive", "archive path")
	logPath  = flag.String("log", "log", "log path")
)

func worker() {
	for i := 1; i < *nworker; i++ {
		go workForever(*workaddr, i, *archPath, *logPath)
	}

	workForever(*workaddr, *nworker, *archPath, *logPath)
}
