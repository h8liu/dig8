package dcrl

import (
	"database/sql"
	"log"
	"path/filepath"
	"time"
)

// InitDB initializes the database
func InitDB(dbPath string) error {
	db, e := sql.Open("sqlite3", dbPath)
	if e != nil {
		return e
	}

	_, e = db.Exec(`create table jobs (
		name text not null primary key,
		archive text not null default "",
		state int not null detault 0,
		total int not null default 0,
		crawled int not null default 0,
		sample text not null default "",
		salt text not null default "",
		worker text not null default "",
		err text not null default "",
		tcreate text not null default "",
		tupdate text not null default "",
		tfinish text not null default ""
	);`)
	if e != nil {
		db.Close()
		return e
	}

	e = db.Close()
	if e != nil {
		return e
	}

	return nil
}

func serve(s *Server) error {
	ticker := time.Tick(time.Minute)

	for {
		select {
		case <-ticker:
			cleanJobs(s)
		case req := <-s.requests:
			switch req.typ {
			case "claimJob":
				worker := req.data.(string)
				j := req.reply.(*JobDesc)
				req.c <- claimJob(s, worker, j)
			case "progress":
				p := req.data.(*Progress)
				hit := req.reply.(*bool)
				req.c <- progress(s, p, hit)
			default:
				log.Printf("error: unknown request %q", req.typ)
			}
		}
	}
}

func ne(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func query(db *sql.DB, q string, args ...interface{}) *sql.Rows {
	rows, e := db.Query(q, args...)
	if e != nil {
		log.Fatal(q)
		log.Fatal(e)
	}
	return rows
}

func claimJob(s *Server, worker string, j *JobDesc) error {
	rows := query(s.db, `
		select name, archive from jobs 
		where state = ? order by tcreate limit 1`,
		State(Created),
	)

	var name, arch string
	for rows.Next() {
		e := rows.Scan(&name, &arch)
		ne(e)
	}

	ne(rows.Err())
	ne(rows.Close())

	if name == "" {
		return nil
	}

	domFile := filepath.Join(s.JobsPath, name)
	doms, e := ReadDomainStrings(domFile)
	ne(e)

	j.Name = name
	j.Archive = arch
	j.Domains = doms

	return nil
}

func progress(s *Server, p *Progress, hit *bool) error {
	return nil
}

func cleanJobs(s *Server) {

}
