package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/rpc"
	"path/filepath"
	"strings"

	"lonnie.io/dig8/dig8"
)

func send() {
	jobName := flag.String("j", "", "job output name")
	serverAddr := flag.String("s", "localhost:5353", "server address")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("error: no input domain list")
	} else if len(args) != 1 {
		log.Fatal("error: multiple input domain lists")
	}

	inputPath := args[0]

	bs, e := ioutil.ReadFile(inputPath)
	ne(e)

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
	if req.Name == "" {
		req.Name = filepath.Base(inputPath)
	}

	c, e := rpc.DialHTTP("tcp", *serverAddr)
	ne(e)

	var reply string
	ne(c.Call("Worker.Crawl", req, &reply))
	if reply != "" {
		log.Print(reply)
	}

	ne(c.Close())
}
