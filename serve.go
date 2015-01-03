package main

import (
	"database/sql"
	"flag"
	"go/build"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var (
	dbPath    = flag.String("db", "jobs.db", "database path")
	doDbInit  = flag.Bool("init", false, "init database")
	serveAddr = flag.String("http", ":8053", "serving address")
)

func dbInit() {
	db, err := sql.Open("sqlite3", *dbPath)
	ne(err)

	q := func(sql string) {
		_, e := db.Exec(sql)
		if e != nil {
			log.Printf("sql: %s\n", sql)
			ne(e)
		}
	}

	q(`create table jobs (
		name text not null primary key,
		state int,
		total int,
		crawled int,
		worker text,
		sample text,
		err text,
		birth text,
		death text
	);`)

	ne(db.Close())
}

func handleApi(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/")

	switch name {
	case "list":
		panic("todo")
	default:
		w.WriteHeader(404)
	}
}

func wwwPath() string {
	pkg, e := build.Import("lonnie.io/dig8", "", build.FindOnly)
	ne(e)
	return filepath.Join(pkg.Dir, "www")
}

func serve() {
	flag.Parse()

	if *doDbInit {
		dbInit()
		return
	}

	http.Handle("/", http.FileServer(http.Dir(wwwPath())))
	http.HandleFunc("/jobs/", handleApi)
	for {
		ne(http.ListenAndServe(*serveAddr, nil))
	}
}
