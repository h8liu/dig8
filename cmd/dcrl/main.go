package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

const helpMsg = `Avaliable commands:
	server  runs the crawler server
	worker  runs a crawler worker
	new     create a new job from a domain list
	deq		dequeues domains from the domain feed queue
	crawl   crawls a single job
	all     runs server, worker and deq all in one process
`

func main() {
	runtime.GOMAXPROCS(4)

	var sub string
	if len(os.Args) > 1 {
		sub = os.Args[1]
		os.Args = os.Args[1:]
	}

	switch sub {
	case "server":
		server()
	case "worker":
		worker()
	case "new":
		newJob()
	case "crawl":
		crawl()
	case "deq":
		dequeue()
	case "all":
		server()
		go func() {
			time.Sleep(time.Second)
			worker()
		}()

		go func() {
			time.Sleep(time.Second)
			dequeue()
		}()
	case "-h":
		fmt.Println(helpMsg)
		os.Exit(0)
	default:
		if sub != "" {
			fmt.Fprintf(os.Stderr, "error: unknown command %q\n", sub)
		}
		fmt.Fprintln(os.Stderr, helpMsg)
		os.Exit(-1)
	}
}
