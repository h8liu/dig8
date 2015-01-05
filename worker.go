package main

import (
	"flag"

	"lonnie.io/dig8/dig8"
)

func worker() {
	archive := flag.String("a", "", "archive prefix")
	serverAddr := flag.String("s", "localhost:5301", "callback address")
	workerAddr := flag.String("w", "localhost:5353", "worker name")
	flag.Parse()

	dig8.WorkerServe(*archive, *serverAddr, *workerAddr)
}
