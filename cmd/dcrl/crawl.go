package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/h8liu/dig8/dcrl"
)

func crawl() {
	name := flag.String("n", "", "job name")
	arch := flag.String("a", "", "archive path")
	db := flag.String("db", "", "database path")
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "missing input")
		os.Exit(-1)
	} else if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "expect only 1 input")
		os.Exit(-1)
	}

	input := args[0]
	jobName := *name
	if jobName == "" {
		jobName = filepath.Base(input)
	}

	doms, e := dcrl.ReadDomains(input)
	if e != nil {
		log.Fatal(e)
	}

	j := &dcrl.Job{
		Name:     jobName,
		Domains:  doms,
		Archive:  *arch,
		DB:       *db,
		Progress: jobProgress,
	}

	e = j.Do()
	if e != nil {
		log.Fatal(e)
	}
}

func jobProgress(p *dcrl.Progress) error {
	log.Println(p.String())
	return nil
}
