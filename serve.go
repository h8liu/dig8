package main

import (
	"flag"
	"go/build"
	"net/http"
	"path/filepath"

	"lonnie.io/dig8/dig8"
)

var (
	dbPath   = flag.String("db", "jobs.db", "database path")
	doDbInit = flag.Bool("init", false, "init database")
	httpAddr = flag.String("http", ":5380", "serving address")
	rpcAddr  = flag.String("rpc", "localhost:5300", "rpc management address")
	cbAddr   = flag.String("cb", "localhost:5301", "callback address")
)

func wwwPath() string {
	pkg, e := build.Import("lonnie.io/dig8", "", build.FindOnly)
	ne(e)
	return filepath.Join(pkg.Dir, "www")
}

// httpAddr (port 5380 default): http interface
// rpcAddr (port 5300 default): rpc interface
// cbAddr (port 5301 default): callback interface
func serve() {
	flag.Parse()

	if *doDbInit {
		dig8.InitDB(*dbPath)
		return
	}

	s, e := dig8.NewServer(*dbPath, *cbAddr)
	ne(e)

	go s.ServeRPC(*rpcAddr)
	go s.ServeCallback()

	http.Handle("/", http.FileServer(http.Dir(wwwPath())))
	http.Handle("/jobs/", s)
	for {
		le(http.ListenAndServe(*httpAddr, nil))
	}
}
