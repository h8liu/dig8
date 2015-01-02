package main

import (
	"flag"

	"lonnie.io/dig8/dig8"
)

func worker() {
	archive := flag.String("a", "", "archive prefix")
	flag.Parse()

	dig8.WorkerServe(*archive)
}
