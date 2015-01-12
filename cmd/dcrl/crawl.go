package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"lonnie.io/dig8/dcrl"
	"lonnie.io/dig8/dns8"
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

	doms, e := readDomains(input)
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

func readDomains(path string) ([]*dns8.Domain, error) {
	fin, e := os.Open(path)
	if e != nil {
		return nil, e
	}

	defer fin.Close()

	s := bufio.NewScanner(fin)
	var ret []*dns8.Domain

	for s.Scan() {
		line := s.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		d, e := dns8.ParseDomain(line)
		if e != nil {
			return nil, e
		}
		ret = append(ret, d)
	}

	e = s.Err()
	if e != nil {
		return nil, e
	}

	return ret, nil
}

func jobProgress(p *dcrl.Progress) {
	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "[%s] ", p.Name)
	if p.Error != "" {
		fmt.Fprintf(buf, "error: %s", p.Error)
	} else if p.Done {
		fmt.Fprintf(buf, "done (%d domains)", p.Total)
	} else {
		fmt.Fprintf(buf, "%d/%d", p.Crawled, p.Total)
	}

	log.Println(buf.String())
}
