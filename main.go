package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"lonnie.io/dig8/dig8"
	"lonnie.io/dig8/dns8"
)

func ne(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func worker() {
	archive := flag.String("a", "", "archive prefix")
	flag.Parse()

	dig8.WorkerServe(*archive)
}

func send() {
	jobName := flag.String("j", "", "job output name")
	serverAddr := flag.String("s", "localhost:5353", "server address")
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

	req := new(dig8.JobRequest)

	doms := strings.Split(string(bs), "\n")
	for _, d := range doms {
		d = strings.TrimSpace(d)
		if d == "" {
			continue
		}
		req.Domains = append(req.Domains, d)
	}

	req.Name = *jobName
	if req.Name == "" {
		req.Name = filepath.Base(inputPath)
	}

	c, e := rpc.DialHTTP("tcp", *serverAddr)
	ne(e)

	var reply string
	ne(c.Call("Worker.Crawl", req, &reply))
	if reply != "" {
		log.Print(reply)
	}

	ne(c.Close())
}

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

func main() {
	runtime.GOMAXPROCS(4)

	var mode string
	if len(os.Args) > 1 {
		mode = os.Args[1]
		os.Args = os.Args[1:]
	}

	switch mode {
	case "worker":
		worker()
	case "send":
		send()
	case "dig":
		dig()
	default:
		fmt.Fprintf(os.Stderr, "error: invalid command %q\n", mode)
		fmt.Fprintf(os.Stderr, "try worker, send or dig\n")
		os.Exit(-1)
	}
}
