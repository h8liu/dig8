package main

import (
	"errors"
	"flag"
	"fmt"
	"net/rpc"

	"lonnie.io/dig8/dcrl"
)

func newJob() {
	tag := flag.String("t", "test", "tag name")
	addr := flag.String("addr", "localhost:5300", "server address")
	arch := flag.String("arch", "default", "archive position")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		ne(errors.New("no input domain list"))
	} else if len(args) != 1 {
		ne(errors.New("expect exactly one domain list"))
	}

	doms, e := dcrl.ReadDomainStrings(args[0])
	ne(e)

	j := &dcrl.NewJobDesc{
		Tag:     *tag,
		Archive: *arch,
		Domains: doms,
	}

	c, e := rpc.DialHTTP("tcp", *addr)
	ne(e)

	var name string
	ne(c.Call("Server.NewJob", j, &name))
	fmt.Println("new job created:", name)

	ne(c.Close())
}
