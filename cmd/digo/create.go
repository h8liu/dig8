package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/rpc"
	"strings"

	"lonnie.io/dig8/digo"
	"lonnie.io/dig8/dns8"
)

func isValidTag(n string) bool {
	if n == "" {
		return false
	}
	for _, r := range n {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		return false
	}
	return true
}

func create() {
	tagName := flag.String("t", "test", "job tag name")
	archive := flag.String("a", "test", "archive position")
	serverAddr := flag.String("s", "localhost:5300", "rpc server address")
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

	lines := strings.Split(string(bs), "\n")
	job := new(digo.NewJob)

	if *tagName == "" {
		log.Fatal("error: empty tag")
	} else if !isValidTag(*tagName) {
		log.Fatal("error: invalid tag %q", *tagName)
	}

	job.Tag = *tagName
	job.Archive = *archive
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		dom, e := dns8.ParseDomain(line)
		if e != nil {
			log.Fatalf("error: %s", e)
		}
		job.Domains = append(job.Domains, dom.String())
	}

	c, e := rpc.DialHTTP("tcp", *serverAddr)
	ne(e)

	var reply string
	ne(c.Call("Server.NewJob", job, &reply))
	if reply != "" {
		log.Print(reply)
	}
	ne(c.Close())
}
