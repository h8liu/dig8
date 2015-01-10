package main

import (
	"flag"
	"fmt"
	"os"

	"lonnie.io/dig8/dns8"
)

func dig() {
	verbose := flag.Bool("v", false, "verbose")
	flag.Parse()

	c, e := dns8.NewClient()
	ne(e)

	t := dns8.NewTerm(c)
	if *verbose {
		t.Log = os.Stdout
	} else {
		t.Log = nil
	}
	t.Out = os.Stdout

	args := flag.Args()
	for _, s := range args {
		d, e := dns8.ParseDomain(s)
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			continue
		}
		fmt.Printf("// %v\n", d)

		_, e = t.T(dns8.NewInfo(d))
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
		}
	}
}
