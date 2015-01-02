package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/rpc"
	"runtime"
	"strings"

	"lonnie.io/dig8/dig8"
)

var (
	jobName    = flag.String("o", "", "job output name")
	inputPath  = flag.String("i", "doms", "input domain list")
	serverAddr = flag.String("s", "localhost:5353", "server address")
	saveAddr   = flag.String("a", "", "archive prefix")
)

func main() {
	runtime.GOMAXPROCS(4)
	flag.Parse()

	if *jobName == "" {
		dig8.WorkerServe(*saveAddr)
		return
	}

	bs, e := ioutil.ReadFile(*inputPath)
	if e != nil {
		log.Fatal(e)
	}

	req := new(dig8.JobRequest)

	doms := strings.Split(string(bs), "\n")
	for _, d := range doms {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		req.Domains = append(req.Domains, d)
	}

	req.Name = *jobName

	c, e := rpc.DialHTTP("tcp", *serverAddr)
	if e != nil {
		log.Fatal(e)
	}

	var reply string
	e = c.Call("Worker.Crawl", req, &reply)
	if e != nil {
		log.Fatal(e)
	} else if reply != "" {
		log.Print(reply)
	}

	e = c.Close()
	if e != nil {
		log.Fatal(e)
	}
}
