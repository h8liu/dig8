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
	case "dig":
		dig()
	default:
		fmt.Fprintf(os.Stderr, "error: invalid command %q\n", mode)
		fmt.Fprintf(os.Stderr, "try worker, send or dig\n")
		os.Exit(-1)
	}
}
