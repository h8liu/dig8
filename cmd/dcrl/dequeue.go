package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"path/filepath"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"lonnie.io/dig8/dcrl"
	"lonnie.io/dig8/dns8"
)

func validArchName(n string) bool {
	if n == "" {
		return false
	}
	for _, r := range n {
		if r >= '0' && r <= '9' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r == '-' || r == '_' {
			continue
		}
		return false
	}

	return true
}

func validTagName(tag string) bool {
	if tag == "" {
		return false
	}
	for _, r := range tag {
		if r >= '0' && r <= '9' {
			continue
		}
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r == '_' {
			continue
		}
		return false
	}

	return true
}

func parseArch(a string) (string, error) {
	if a == "" {
		return a, nil
	}

	names := strings.Split(a, ".")

	for _, n := range names {
		if !validArchName(n) {
			return a, fmt.Errorf("invalid archive path: %s", a)
		}
	}

	return filepath.Join(names...), nil
}

func parseLine(line string) (name, arch string) {
	sep := strings.LastIndex(line, " ")
	if sep == -1 {
		return line, ""
	}
	return line[:sep], line[sep+1:]
}

func dequeue() {
	src := flag.String("src", "ccied50.sysnet.ucsd.edu:6379", "source address")
	qname := flag.String("q", "domains", "the queue reading")
	dest := flag.String("dest", "localhost:5300", "rpc server address")
	batch := flag.Int("b", 10000, "batching size")
	tag := flag.String("tag", "feed", "tag name")
	flag.Parse()

	if !validTagName(*tag) {
		ne(fmt.Errorf("invalid tag name: %q", *tag))
	} else if *batch <= 0 {
		ne(fmt.Errorf("invalid batch size: %d", *batch))
	}

	c, e := redis.Dial("tcp", *src)
	ne(e)

	wasSleeping := false

	for {
		n, e := redis.Uint64(c.Do("llen", *qname))
		ne(e)

		if n == 0 {
			if !wasSleeping {
				log.Print("queue empty, go to sleep now.\n")
				wasSleeping = true
				time.Sleep(time.Minute)
			}
			continue
		} else {
			wasSleeping = false
			log.Printf("%d domains in the queue\n", n)
		}

		lines, e := redis.Strings(c.Do("lrange", *qname, -*batch, -1))
		ne(e)
		log.Printf("%d lines read\n", len(lines))

		set := make(map[string][]string)
		dedup := make(map[string]struct{})

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			name, arch := parseLine(line)
			d, e := dns8.ParseDomain(name)
			if e != nil {
				log.Print(line, e)
				continue
			}
			name = d.String() // regulate the domain

			arch, e = parseArch(arch)
			if e != nil {
				log.Print(line, e)
				continue
			}

			key := name + " " + arch
			if _, found := dedup[key]; found {
				continue
			}

			set[arch] = append(set[arch], name)
			dedup[key] = struct{}{}
		}

		server, e := rpc.DialHTTP("tcp", *dest)
		ne(e)

		total := 0
		for arch, doms := range set {
			desc := &dcrl.NewJobDesc{
				Tag:     *tag,
				Archive: arch,
				Domains: doms,
			}

			var jobName string
			ne(server.Call("Server.NewJob", desc, &jobName))
			n := len(doms)
			log.Printf("[%s] new job with %d domains\n", jobName, n)
			total += n
		}

		ne(server.Close())

		_, e = c.Do("ltrim", *qname, 0, -*batch)
		ne(e)
		log.Printf("%d domains scheduled with %d jobs", total, len(set))
	}

	// should be unreachable
	ne(c.Close())
}
