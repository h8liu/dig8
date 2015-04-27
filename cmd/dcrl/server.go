package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"lonnie.io/dig8/dcrl"
)

func ne(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(-1)
	}
}

var (
	dbPath   = flag.String("db", "jobs.db", "database path")
	doInit   = flag.Bool("init", false, "init database")
	jobsPath = flag.String("jobs", "jobs", "jobs path")
	addr     = flag.String("addr", ":5300", "serving address")
)

func server() {
	if *doInit {
		ne(dcrl.InitDB(*dbPath))
		return
	}

	ne(os.MkdirAll(*jobsPath, 0775))

	s := &dcrl.Server{
		JobsDB:   *dbPath,
		JobsPath: *jobsPath,
	}

	ne(dcrl.Serve(s)) // listen on the request channel

	rs := rpc.NewServer()
	ne(rs.RegisterName("Server", (*dcrl.RPCServer)(s)))

	c, e := net.Listen("tcp", *addr)
	ne(e)

	log.Printf("serving at %s", *addr)
	for {
		e = http.Serve(c, rs)
		if e != nil {
			log.Print(e)
		}
	}
}
