package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"

	"lonnie.io/dig8/dig8"
	"lonnie.io/dig8/dns8"
)

func isValidName(n string) bool {
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

func parseArch(a string) (string, error) {
	if a == "" {
		return a, nil
	}

	names := strings.Split(a, ".")

	for _, n := range names {
		if !isValidName(n) {
			return a, fmt.Errorf("invalid tag: %s", a)
		}
	}

	return strings.Join(names, "/"), nil
}

const batchSize = 10000

// pumps domains out of the CESR redis domain queue
func pump() {
	testing := flag.Bool("t", false, "testing with localhost")
	serverAddr := flag.String("s", "localhost:5300", "rpc server address")
	flag.Parse()

	addr := "ccied50.sysnet.ucsd.edu:6379"
	if *testing {
		addr = "localhost:6379"
	}

	c, e := redis.Dial("tcp", addr)
	ne(e)

	var wasSleeping bool
	for {
		n, e := redis.Uint64(c.Do("llen", "domains"))
		ne(e)

		if n == 0 {
			if !wasSleeping {
				log.Printf("queue empty\n", n)
				log.Printf("sleeping")
				wasSleeping = true
				time.Sleep(time.Minute)
			}
			continue
		} else {
			wasSleeping = false
			log.Printf("%d domains in the queue now\n", n)
		}

		lines, e := redis.Strings(c.Do("lrange", "domains", -batchSize, -1))
		ne(e)
		log.Printf("%d lines read from queue\n", len(lines))

		set := make(map[string][]string)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			sep := strings.LastIndex(line, " ")
			var name, arch string
			if sep == -1 {
				name = line
			} else {
				name = line[:sep]
				arch = line[sep+1:]
			}

			d, e := dns8.ParseDomain(name)
			if e != nil {
				log.Print(line, e) // ignore
				continue
			}
			name = d.String()

			arch, e = parseArch(arch)
			if e != nil {
				log.Print(line, e) // ignore
				continue
			}

			set[arch] = append(set[arch], name)
		}

		server, e := rpc.DialHTTP("tcp", *serverAddr)
		ne(e)

		total := 0
		for arch, doms := range set {
			job := &dig8.NewJob{
				Tag:     "feed",
				Archive: arch,
				Domains: doms,
			}

			var reply string
			ne(server.Call("Server.NewJob", job, &reply))
			if reply != "" {
				ne(errors.New(reply))
			}

			log.Printf("%s: %d domains\n", arch, len(doms))
			total += len(doms)
		}

		ne(server.Close())

		_, e = c.Do("ltrim", "domains", 0, -batchSize)
		ne(e)
		log.Printf("%d domains scheduled with %d jobs", total, len(set))
	}

	ne(c.Close())
}
