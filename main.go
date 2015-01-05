package main

import (
	"fmt"
	"os"
	"runtime"
)

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
	case "create":
		create()
	case "dig":
		dig()
	case "serve":
		serve()
	case "pump":
		pump()
	default:
		if mode != "" {
			fmt.Fprintf(os.Stderr, "error: invalid command %q\n", mode)
		}
		fmt.Fprintf(os.Stderr, helpMsg)
		os.Exit(-1)
	}
}

const helpMsg = `Available commands:
    worker  launch a dns crawling worker
    send    send a job to a worker
    create  create a job and send it to server
    dig     test a domain
    serve   launch a dns crawling master
    pump    pump domains from the domain queue
`
