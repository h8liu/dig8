package dcrl

import (
	"crypto/rand"
	"database/sql"
	"encoding/base32"
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
		state int not null default 0,
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

	return db.Close()
}

// Serve launches the server
func Serve(s *Server) error {
	var e error
	s.db, e = sql.Open("sqlite3", s.JobsDB)
	if e != nil {
		return e
	}

	s.requests = make(chan *request, 64)

	go serve(s)
	return nil
}

func serve(s *Server) error {
	defer s.db.Close()

	ticker := time.Tick(time.Minute * 3)
	// ticker := time.Tick(time.Second * 2)
	progLast := make(map[string]int)

	for {
		select {
		case <-ticker:
			cleanJobs(s)
		case req := <-s.requests:
			switch req.typ {
			case "claimJob":
				worker := req.data.(string)
				j := req.reply.(*JobDesc)
				e := claimJob(s, worker, j)
				if e == nil {
					if j.Name != "" {
						log.Printf("[%s] claimed by %s", j.Name, worker)
					}
				} else {
					log.Print(e)
				}
				req.c <- e
			case "progress":
				p := req.data.(*Progress)
				okay := req.reply.(*bool)

				if p.Error == "" && !p.Done {
					last, found := progLast[p.Name]
					if !found || last != p.Crawled {
						log.Print(p)
					}
					progLast[p.Name] = p.Crawled
				} else {
					log.Print(p)
				}
				req.c <- progress(s, p, okay)
			case "newJob":
				j := req.data.(*NewJobDesc)
				name := req.reply.(*string)
				e := newJob(s, j, name)
				if e == nil {
					log.Printf("[%s] created: %d domains",
						*name, len(j.Domains),
					)
				} else {
					log.Print(e)
				}
				req.c <- e
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

var enc = base32.NewEncoding(
	"0123456789abcdefghijklmnopqrstuv",
)

func randName() string {
	var buf [8]byte
	_, e := rand.Read(buf[:])
	ne(e)
	s := enc.EncodeToString(buf[:])
	if len(s) < 6 {
		panic("bug")
	}
	return s[:6]
}

func countRows(rows *sql.Rows) int {
	cnt := 0
	for rows.Next() {
		cnt++
	}
	ne(rows.Err())
	ne(rows.Close())
	return cnt
}

func newName(s *Server, tag string) string {
	for {
		name := tag + "." + randName()
		rows := query(s.db, `select name from jobs where name = ?`, name)
		if countRows(rows) == 0 {
			return name
		}
	}
}

func newJob(s *Server, j *NewJobDesc, name *string) error {
	*name = newName(s, j.Tag)

	// TODO: save domain to file
	domFile := filepath.Join(s.JobsPath, *name)
	ne(WriteDomainStrings(domFile, j.Domains))

	exec(s.db, `
		insert into jobs
		(name, archive, state, total, crawled, tcreate)
		values
		(?, ?, ?, ?, ?, ?)`,
		*name, j.Archive, int(Created), len(j.Domains), 0, timeNow(),
	)

	return nil
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
		j.Name = ""
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

func progress(s *Server, p *Progress, okay *bool) error {
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

	*okay = true

	return nil
}

func cleanJobs(s *Server) {
	tago := timeAgo(time.Minute * 3)
	tnow := timeNow()

	// restart the errrored ones
	exec(s.db, `
		update jobs set
		state = ?, tupdate = ?,
		worker = "", crawled = 0, err = "", retried = retried + 1
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
		int(Errored),
		int(Crawling), tago,
	)
}
