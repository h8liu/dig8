package main

import (
	"fmt"
	"os"
	"runtime"
)

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

const helpMsg = `Avaliable commands:
	server  runs the crawler server
	worker  runs a crawler worker
	new     create a new job from a domain list
	crawl   crawls a single job
`
