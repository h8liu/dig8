package main

import (
	"flag"

	"lonnie.io/dig8/digo"
)

func worker() {
	archive := flag.String("a", "", "archive prefix")
	serverAddr := flag.String("s", "localhost:5301", "callback address")
	workerAddr := flag.String("w", "localhost:5353", "worker name")
	listenAddr := flag.String("l", "localhost:5353", "listen address")
	flag.Parse()

	digo.WorkerServe(*archive, *serverAddr, *workerAddr, *listenAddr)
}
