package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/h8liu/dig8/dcrl"
)

var (
	arch = flag.String("a", "", "archive path")
	db = flag.String("db", "", "database path")
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		log.Fatalln("missing input")
	} else if len(args) != 1 {
		log.Fatalln("expect only 1 input")
	}

	input := args[0]
	jobName := filepath.Base(input)
	doms, e := dcrl.ReadDomains(input)
	if e != nil {
		log.Fatalln(e)
	}

	j := &dcrl.Job{
		Name:     jobName,
		Domains:  doms,
		Archive:  *arch,
		DB:       *db,
		Progress: func (p *dcrl.Progress) error {
			log.Println(p.String())
			return nil
		},
	}

	e = j.Do()
	if e != nil {
		log.Fatal(e)
	}
}
