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
		total int not null default 0,
		sample text not null default "",
		salt text not null default "",
		state int not null detault 0,
		worker text not null default "",
		crawled int not null default 0,
		retried int not null default 0,
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

// Serve launches the server
func Serve(s *Server) error {
	var e error
	s.db, e = sql.Open("sqlite3", s.JobsDB)
	if e != nil {
		return e
	}
	defer s.db.Close()

	s.requests = make(chan *request, 64)

	return serve(s)
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
		log.Print(q)
		log.Fatal(e)
	}
	return rows
}

func exec(db *sql.DB, q string, args ...interface{}) sql.Result {
	res, e := db.Exec(q, args...)
	if e != nil {
		log.Print(q)
		log.Fatal(e)
	}
	return res
}

const tfmt = "2006-01-02 15:04:05"

func timeNow() string {
	t := time.Now().UTC()
	return t.Format(tfmt)
}

func timeAgo(d time.Duration) string {
	t := time.Now().Add(-d).UTC()
	return t.Format(tfmt)
}

func claimJob(s *Server, worker string, j *JobDesc) error {
	rows := query(s.db, `
		select name, archive from jobs 
		where state = ? order by tcreate limit 1`,
		int(Created),
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

	exec(s.db, `
		update jobs set
		worker = ?, state = ?, tupdate = ?,
		crawled = 0, err = "", tfinish = ""
		where name = ?`,
		worker, int(Crawling), timeNow(),
		name,
	)

	return nil
}

func progress(s *Server, p *Progress, hit *bool) error {
	tnow := timeNow()
	if p.Error != "" {
		exec(s.db, `
			update jobs set
			state = ?, tupdate = ?, err = ?, tfinish = ?
			where name = ?`,
			int(Errored), tnow, p.Error, tnow, p.Name,
		)
	} else if p.Done {
		exec(s.db, `
			update jobs set
			state = ?, tupdate = ?, tfinish = ?,
			crawled = total, err = ""
			where name = ?`,
			int(Done), tnow, tnow, p.Name,
		)
	} else {
		exec(s.db, `
			update jobs set
			state = ?, tupdate = ?, crawled = ?,
			err = ""
			where name = ?`,
			int(Crawling), tnow, p.Crawled, p.Name,
		)
	}

	return nil
}

func cleanJobs(s *Server) {
	tago := timeAgo(time.Minute * 3)
	tnow := timeNow()

	// restart the errrored ones
	exec(s.db, `
		update jobs set
		state = ?, tupdate = ?
		worker = "", crawled = 0, err = "", retried = retried + 1,
		where state = ? and tfinish < ? and retried < 3`,
		int(Created), tnow,
		int(Errored), tago,
	)

	// archive the finished ones
	exec(s.db, `
		update jobs set
		state = ?
		where state = ? and tfinish < ?`,
		int(Archived),
		int(Done), tago,
	)

	// error the lost crawling ones
	exec(s.db, `
		update jobs set
		state = ?, err = "worker lost"
		where state = ? and tupdate < ?`,
		int(Crawling),
		int(Errored), tago,
	)
}
