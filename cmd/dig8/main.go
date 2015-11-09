package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/h8liu/dig8/dns8"
)

func ne(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func main() {
	quiet := flag.Bool("q", false, "quiet")
	flag.Parse()

	c, e := dns8.NewClient()
	ne(e)

	t := dns8.NewTerm(c)
	if !*quiet {
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
